package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"
)

type Conversation struct {
	ID           string `json:"id"` // "user-123" or "group-456"
	Name         string `json:"name"`
	Type         string `json:"type"` // "user" or "group"
	UnreadCount  int    `json:"unread_count"`
}

// GetConversationsHandler will fetch all conversations for the current user
func (app *App) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations := make([]Conversation, 0)
	existingUsers := make(map[int64]bool)

	// Get group conversations
	rows, err := app.DB.Query(`
		SELECT g.id, g.title
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND gm.status = 'accepted'
	`, currentUserID)
	if err != nil {
		http.Error(w, "Failed to get group conversations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var groupID int64
		var groupTitle string
		if err := rows.Scan(&groupID, &groupTitle); err != nil {
			http.Error(w, "Failed to scan group conversation", http.StatusInternalServerError)
			return
		}
		// Unread count for groups is not implemented in this version.
		conversations = append(conversations, Conversation{
			ID:          "group-" + strconv.FormatInt(groupID, 10),
			Name:        groupTitle,
			Type:        "group",
			UnreadCount: 0,
		})
	}
	rows.Close()

	// Get user conversations from follower relationships
	userConversationsQuery := `
		SELECT u.id, u.first_name, u.last_name FROM users u JOIN followers f ON u.id = f.followed_id WHERE f.follower_id = ? AND f.status = 'accepted'
		UNION
		SELECT u.id, u.first_name, u.last_name FROM users u JOIN followers f ON u.id = f.follower_id WHERE f.followed_id = ? AND f.status = 'accepted'
	`
	rows, err = app.DB.Query(userConversationsQuery, currentUserID, currentUserID)
	if err != nil {
		http.Error(w, "Failed to get user conversations from followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var firstName, lastName string
		if err := rows.Scan(&userID, &firstName, &lastName); err != nil {
			http.Error(w, "Failed to scan user conversation", http.StatusInternalServerError)
			return
		}
		if _, exists := existingUsers[userID]; !exists {
			var unreadCount int
			err := app.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE sender_id = ? AND receiver_id = ? AND is_read = FALSE", userID, currentUserID).Scan(&unreadCount)
			if err != nil {
				http.Error(w, "Failed to get unread count for user", http.StatusInternalServerError)
				return
			}

			conversations = append(conversations, Conversation{
				ID:          "user-" + strconv.FormatInt(userID, 10),
				Name:        firstName + " " + lastName,
				Type:        "user",
				UnreadCount: unreadCount,
			})
			existingUsers[userID] = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// ... (rest of the file remains the same)
// GetMessagesHandler, scanMessages, GetUnreadMessageCountHandler
// ... (rest of the file remains the same)
func (app *App) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	userIDStr := query.Get("user_id")
	groupIDStr := query.Get("group_id")

	var rows *sql.Rows
	var err error

	baseQuery := `
		SELECT m.id, m.sender_id, u.first_name, m.receiver_id, m.group_id, m.content, m.created_at, m.is_read
		FROM messages m
		JOIN users u ON m.sender_id = u.id
	`

	if userIDStr != "" {
		otherUserID, _ := strconv.ParseInt(userIDStr, 10, 64)
		rows, err = app.DB.Query(baseQuery + `
			WHERE m.group_id IS NULL AND ((m.sender_id = ? AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = ?))
			ORDER BY m.created_at ASC
		`, currentUserID, otherUserID, otherUserID, currentUserID)
	} else if groupIDStr != "" {
		groupID, _ := strconv.ParseInt(groupIDStr, 10, 64)
		var status string
		err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, currentUserID).Scan(&status)
		if err != nil || status != "accepted" {
			http.Error(w, "Not a member of this group", http.StatusForbidden)
			return
		}

		rows, err = app.DB.Query(baseQuery + `WHERE m.group_id = ? ORDER BY m.created_at ASC`, groupID)
	} else {
		http.Error(w, "Missing user_id or group_id parameter", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Failed to execute messages query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		http.Error(w, "Failed to scan messages", http.StatusInternalServerError)
		return
	}

	// After fetching, mark these messages as read for the current user
	if userIDStr != "" {
		otherUserID, _ := strconv.ParseInt(userIDStr, 10, 64)
		_, err := app.DB.Exec("UPDATE messages SET is_read = TRUE WHERE receiver_id = ? AND sender_id = ? AND is_read = FALSE", currentUserID, otherUserID)
		if err != nil {
			// Log error but don't fail the request, as the user has received their messages
			log.Printf("Failed to mark messages as read: %v", err)
		}
	}
	// Note: Group message read status is not implemented in this version.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var receiverID sql.NullInt64
		var groupID sql.NullInt64
		var isRead sql.NullBool

		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.SenderName,
			&receiverID,
			&groupID,
			&msg.Content,
			&msg.CreatedAt,
			&isRead,
		)
		if err != nil {
			return nil, err
		}
		if receiverID.Valid {
			msg.ReceiverID = &receiverID.Int64
		}
		if groupID.Valid {
			msg.GroupID = &groupID.Int64
		}
		if isRead.Valid {
			msg.IsRead = &isRead.Bool
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (app *App) GetUnreadMessageCountHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var unreadCount int
	err := app.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE receiver_id = ? AND is_read = FALSE", currentUserID).Scan(&unreadCount)
	if err != nil {
		http.Error(w, "Failed to get unread message count", http.StatusInternalServerError)
		return
	}

	response := map[string]int{"unread_count": unreadCount}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// MarkAllMessagesAsReadHandler marks all unread messages for the current user as read.
func (app *App) MarkAllMessagesAsReadHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE messages SET is_read = TRUE WHERE receiver_id = ? AND is_read = FALSE")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(currentUserID)
	if err != nil {
		http.Error(w, "Failed to mark all messages as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All messages marked as read"))
}
