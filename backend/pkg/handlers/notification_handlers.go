package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social-network/backend/pkg/models"
	"social-network/backend/pkg/router"


)

func (app *App) createNotification(userID int64, notifType string, message string, fromUserID int64, relatedID int64) {
	var sender models.User
	err := app.DB.QueryRow("SELECT first_name, last_name FROM users WHERE id = ?", fromUserID).Scan(&sender.FirstName, &sender.LastName)
	if err != nil {
		log.Printf("Failed to get sender info for notification: %v", err)
		return
	}

	fullMessage := fmt.Sprintf("%s %s %s", sender.FirstName, sender.LastName, message)

	stmt, err := app.DB.Prepare("INSERT INTO notifications (user_id, type, message, related_id) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare notification statement: %v", err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, notifType, fullMessage, relatedID)
	if err != nil {
		log.Printf("Failed to create notification: %v", err)
		return
	}

	// Get the newly created notification to send over WebSocket
	notifID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert id for notification: %v", err)
		return
	}

	var notif models.Notification
	err = app.DB.QueryRow("SELECT id, user_id, type, message, is_read, related_id, created_at FROM notifications WHERE id = ?", notifID).Scan(&notif.ID, &notif.UserID, &notif.Type, &notif.Message, &notif.IsRead, &notif.RelatedID, &notif.CreatedAt)
	if err != nil {
		log.Printf("Failed to get new notification for websocket: %v", err)
		return
	}

	// Send notification via WebSocket
	app.Hub.RouteNotification(&notif)
}

func (app *App) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	rows, err := app.DB.Query("SELECT id, user_id, type, message, is_read, related_id, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Type, &notification.Message, &notification.IsRead, &notification.RelatedID, &notification.CreatedAt); err != nil {
			http.Error(w, "Failed to scan notification", http.StatusInternalServerError)
			return
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		http.Error(w, "Failed to encode notifications", http.StatusInternalServerError)
	}
}

func (app *App) MarkNotificationAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	notificationIDStr := router.ForContext(r.Context(), "id")
	if notificationIDStr == "" {
		http.Error(w, "Notification ID is missing", http.StatusBadRequest)
		return
	}

	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE notifications SET is_read = TRUE WHERE id = ? AND user_id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(notificationID, userID)
	if err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Notification not found or you are not authorized to mark it as read", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification marked as read"))
}

func (app *App) DeleteNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	notificationIDStr := router.ForContext(r.Context(), "id")
	if notificationIDStr == "" {
		http.Error(w, "Notification ID is missing", http.StatusBadRequest)
		return
	}

	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("DELETE FROM notifications WHERE id = ? AND user_id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(notificationID, userID)
	if err != nil {
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Notification not found or you are not authorized to delete it", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification deleted successfully"))
}

func (app *App) GetUnreadNotificationCountHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var unreadCount int
	err := app.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0", userID).Scan(&unreadCount)
	if err != nil {
		http.Error(w, "Failed to get unread notification count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"unread_count": unreadCount}); err != nil {
		http.Error(w, "Failed to encode unread notification count", http.StatusInternalServerError)
	}
}

func (app *App) MarkAllNotificationsAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE notifications SET is_read = TRUE WHERE user_id = ? AND is_read = FALSE")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		http.Error(w, "Failed to mark all notifications as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All notifications marked as read"))
}
