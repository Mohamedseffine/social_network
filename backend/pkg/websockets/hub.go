package websockets

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/backend/pkg/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte // Buffered channel of outbound messages.
	userID int64
}

// IncomingMessage defines the structure of messages received from clients
type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload MessagePayload  `json:"payload"`
}

// MessagePayload is the actual content of the message
type MessagePayload struct {
	Content     string `json:"content"`
	RecipientID int64  `json:"recipient_id,omitempty"`
	GroupID     int64  `json:"group_id,omitempty"`
}

// OutgoingMessage defines the structure of messages sent to clients
type OutgoingMessage struct {
	Content    string `json:"content"`
	SenderID   int64  `json:"sender_id"`
	SenderName string `json:"sender_name"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"` // "private" or "group"
	TargetID   int64  `json:"target_id"` // UserID for private, GroupID for group
}


func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Add sender information and timestamp
		now := time.Now()

		// Persist message to database
		stmt, err := c.hub.db.Prepare("INSERT INTO messages (sender_id, receiver_id, group_id, content, created_at) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Failed to prepare statement for message insert: %v", err)
			continue
		}

		var receiverID, groupID sql.NullInt64
		if incomingMsg.Payload.RecipientID != 0 {
			receiverID = sql.NullInt64{Int64: incomingMsg.Payload.RecipientID, Valid: true}
		}
		if incomingMsg.Payload.GroupID != 0 {
			groupID = sql.NullInt64{Int64: incomingMsg.Payload.GroupID, Valid: true}
		}

		_, err = stmt.Exec(c.userID, receiverID, groupID, incomingMsg.Payload.Content, now)
		stmt.Close()
		if err != nil {
			log.Printf("Failed to insert message into database: %v", err)
			continue
		}

		// Get sender's name for the outgoing message
		var senderName string
		err = c.hub.db.QueryRow("SELECT first_name FROM users WHERE id = ?", c.userID).Scan(&senderName)
		if err != nil {
			log.Printf("Failed to get sender name: %v", err)
			senderName = "Unknown"
		}

		// Create the message to be routed
		outgoingMsg := OutgoingMessage{
			Content:    incomingMsg.Payload.Content,
			SenderID:   c.userID,
			SenderName: senderName,
			Timestamp:  now.Format(time.RFC3339),
			Type:       incomingMsg.Type,
			TargetID:   incomingMsg.Payload.RecipientID,
		}
		if incomingMsg.Type == "group_message" {
			outgoingMsg.TargetID = incomingMsg.Payload.GroupID
		}

		c.hub.route <- &outgoingMsg

		// Notifications for new messages are handled by the unread count, not the main notification system.
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}

// WebSocketMessage is the top-level structure for all messages.
type WebSocketMessage struct {
	Type    string      `json:"type"` // "chat_message", "notification", "error", etc.
	Payload interface{} `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients        map[int64]*Client // Map userID to Client
	route          chan *OutgoingMessage
	register       chan *Client
	unregister     chan *Client
	mu             sync.Mutex
	db             *sql.DB
}

func (h *Hub) RouteNotification(notification *models.Notification) {
	msg := WebSocketMessage{
		Type:    "new_notification",
		Payload: notification,
	}
	jsonMsg, _ := json.Marshal(msg)

	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[notification.UserID]; ok {
		select {
		case client.send <- jsonMsg:
		default:
			close(client.send)
			delete(h.clients, client.userID)
		}
	}
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		route:      make(chan *OutgoingMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.route:
			h.mu.Lock()

			wrappedMsg := WebSocketMessage{
				Type: "chat_message",
				Payload: message,
			}
			jsonMsg, _ := json.Marshal(wrappedMsg)

			if message.Type == "private_message" {
				// Send to recipient
				if recipientClient, ok := h.clients[message.TargetID]; ok {
					select {
					case recipientClient.send <- jsonMsg:
					default:
						close(recipientClient.send)
						delete(h.clients, recipientClient.userID)
					}
				}
				// Send back to sender for UI sync
				if senderClient, ok := h.clients[message.SenderID]; ok {
					select {
					case senderClient.send <- jsonMsg:
					default:
						close(senderClient.send)
						delete(h.clients, senderClient.userID)
					}
				}
			} else if message.Type == "group_message" {
				rows, err := h.db.Query("SELECT user_id FROM group_members WHERE group_id = ? AND status = 'accepted'", message.TargetID)
				if err != nil {
					log.Printf("Failed to get group members: %v", err)
					continue
				}
				defer rows.Close()

				for rows.Next() {
					var memberID int64
					if err := rows.Scan(&memberID); err != nil {
						log.Printf("Failed to scan group member: %v", err)
						continue
					}
					if memberClient, ok := h.clients[memberID]; ok {
						select {
						case memberClient.send <- jsonMsg:
						default:
							close(memberClient.send)
							delete(h.clients, memberID)
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: userID}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	// Notify user of successful connection
	msg := map[string]string{"type": "connection", "status": "successful"}
	jsonMsg, _ := json.Marshal(msg)
	client.send <- jsonMsg
}
