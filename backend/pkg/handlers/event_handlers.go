package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"social-network/backend/pkg/models"

	"github.com/gorilla/mux"
)

type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventTime   string `json:"event_time"`
}

func (app *App) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can create events", http.StatusForbidden)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		http.Error(w, "Invalid event time format", http.StatusBadRequest)
		return
	}
	if eventTime.Before(time.Now()) {
		http.Error(w, "Event time must be in the future", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("INSERT INTO events (group_id, creator_id, title, description, event_time) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(groupID, userID, req.Title, req.Description, eventTime)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}
	eventID, err := res.LastInsertId()
	if err != nil {
		log.Printf("CRITICAL: Failed to get last insert ID for event: %v", err)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Event created successfully but failed to start notification process"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Event created successfully"))

	// Run notification creation in a separate goroutine
	go func() {
		log.Printf("DEBUG: Event created with ID %d. Starting notification process for group %d.", eventID, groupID)

		rows, err := app.DB.Query("SELECT user_id FROM group_members WHERE group_id = ? AND user_id != ? AND status = 'accepted'", groupID, userID)
		if err != nil {
			log.Printf("DEBUG: Failed to get group members for notification: %v", err)
			return
		}
		defer rows.Close()

		var memberIDs []int64
		for rows.Next() {
			var memberID int64
			if err := rows.Scan(&memberID); err != nil {
				log.Printf("DEBUG: Failed to scan member ID for notification: %v", err)
				continue
			}
			memberIDs = append(memberIDs, memberID)
		}
		log.Printf("DEBUG: Found %d accepted members to notify: %v", len(memberIDs), memberIDs)

		if len(memberIDs) > 0 {
			var group models.Group
			err = app.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&group.Title)
			if err != nil {
				log.Printf("DEBUG: Failed to get group info for notification: %v", err)
				return // Can't proceed without group title
			}
			log.Printf("DEBUG: Group title for notification: %s", group.Title)

			for _, memberID := range memberIDs {
				log.Printf("DEBUG: Creating notification for member ID: %d", memberID)
				message := fmt.Sprintf("created a new event in '%s': %s", group.Title, req.Title)
				app.createNotification(memberID, "new_group_event", message, userID, eventID)
			}
		}
	}()
}

func (app *App) GetGroupEventsHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can view group events", http.StatusForbidden)
		return
	}

	rows, err := app.DB.Query("SELECT id, group_id, creator_id, title, description, event_time, created_at FROM events WHERE group_id = ? ORDER BY event_time DESC", groupID)
	if err != nil {
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.GroupID, &event.CreatorID, &event.Title, &event.Description, &event.EventTime, &event.CreatedAt); err != nil {
			http.Error(w, "Failed to scan event", http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode events", http.StatusInternalServerError)
	}
}

type RespondToEventRequest struct {
	Status string `json:"status"`
}

func (app *App) RespondToEventHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	eventIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Event ID is missing", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get the group id from the event id
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM events WHERE id = ?", eventID).Scan(&groupID)
	if err != nil {
		http.Error(w, "Failed to get group ID from event ID", http.StatusInternalServerError)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can respond to events", http.StatusForbidden)
		return
	}

	var req RespondToEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "going" && req.Status != "not_going" {
		http.Error(w, "Invalid status. Must be 'going' or 'not_going'", http.StatusBadRequest)
		return
	}

	// Check if the user has already responded
	var attendeeID int64
	err = app.DB.QueryRow("SELECT id FROM event_attendees WHERE event_id = ? AND user_id = ?", eventID, userID).Scan(&attendeeID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check attendance status", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// Insert new response
		stmt, err := app.DB.Prepare("INSERT INTO event_attendees (event_id, user_id, status) VALUES (?, ?, ?)")
		if err != nil {
			http.Error(w, "Failed to prepare insert statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(eventID, userID, req.Status)
		if err != nil {
			http.Error(w, "Failed to respond to event", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing response
		stmt, err := app.DB.Prepare("UPDATE event_attendees SET status = ? WHERE id = ?")
		if err != nil {
			http.Error(w, "Failed to prepare update statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(req.Status, attendeeID)
		if err != nil {
			http.Error(w, "Failed to update response", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Response recorded"))
}

func (app *App) GetEventAttendeesHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	eventIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Event ID is missing", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get the group id from the event id
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM events WHERE id = ?", eventID).Scan(&groupID)
	if err != nil {
		http.Error(w, "Failed to get group ID from event ID", http.StatusInternalServerError)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can view event attendees", http.StatusForbidden)
		return
	}

	rows, err := app.DB.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.nickname, ea.status
		FROM users u
		JOIN event_attendees ea ON u.id = ea.user_id
		WHERE ea.event_id = ?
	`, eventID)
	if err != nil {
		http.Error(w, "Failed to get event attendees", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Attendee struct {
		models.User
		Status string `json:"status"`
	}

	var attendees []Attendee
	for rows.Next() {
		var attendee Attendee
		if err := rows.Scan(&attendee.ID, &attendee.Email, &attendee.FirstName, &attendee.LastName, &attendee.Avatar, &attendee.Nickname, &attendee.Status); err != nil {
			http.Error(w, "Failed to scan attendee", http.StatusInternalServerError)
			return
		}
		attendees = append(attendees, attendee)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over attendees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(attendees); err != nil {
		http.Error(w, "Failed to encode attendees", http.StatusInternalServerError)
	}
}
