package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (app *App) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	followedIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	followedID, err := strconv.ParseInt(followedIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if currentUserID == followedID {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	// Check if the user to be followed exists and get their profile privacy
	var isPublic bool
	err = app.DB.QueryRow("SELECT profile_is_public FROM users WHERE id = ?", followedID).Scan(&isPublic)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Check if a follow request already exists
	var status string
	err = app.DB.QueryRow("SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?", currentUserID, followedID).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check follow status", http.StatusInternalServerError)
		return
	}
	if err == nil {
		http.Error(w, "You are already following or have a pending request for this user", http.StatusBadRequest)
		return
	}

	// If the profile is public, the follow is accepted immediately. Otherwise, it's pending.
	followStatus := "pending"
	if isPublic {
		followStatus = "accepted"
	}

	stmt, err := app.DB.Prepare("INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(currentUserID, followedID, followStatus)
	if err != nil {
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if followStatus == "pending" {
		w.Write([]byte("Follow request sent"))

		// Create notification for follow request
		followRequestID, err := res.LastInsertId()
		if err != nil {
			log.Printf("Failed to get last insert ID for follow request: %v", err)
			return
		}
		app.createNotification(followedID, "follow_request", fmt.Sprintf("wants to follow you."), currentUserID, followRequestID)

	} else {
		w.Write([]byte("User followed successfully"))
		// Create notification for new follower
		app.createNotification(followedID, "new_follower", fmt.Sprintf("is now following you."), currentUserID, 0)
	}
}

func (app *App) UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	followedIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	followedID, err := strconv.ParseInt(followedIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("DELETE FROM followers WHERE follower_id = ? AND followed_id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(currentUserID, followedID)
	if err != nil {
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "You are not following this user", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User unfollowed successfully"))
}

func (app *App) AcceptFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Request ID not provided", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE followers SET status = 'accepted' WHERE id = ? AND followed_id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(requestID, currentUserID)
	if err != nil {
		http.Error(w, "Failed to accept follow request", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Follow request not found or you are not authorized to accept it", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Follow request accepted"))
}

func (app *App) DeclineFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Request ID not provided", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("DELETE FROM followers WHERE id = ? AND followed_id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(requestID, currentUserID)
	if err != nil {
		http.Error(w, "Failed to decline follow request", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Follow request not found or you are not authorized to decline it", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Follow request declined"))
}

func (app *App) GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rows, err := app.DB.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.nickname
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.followed_id = ? AND f.status = 'accepted'
	`, userID)
	if err != nil {
		http.Error(w, "Failed to get followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	followers := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Avatar, &user.Nickname); err != nil {
			http.Error(w, "Failed to scan follower", http.StatusInternalServerError)
			return
		}
		followers = append(followers, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (app *App) GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rows, err := app.DB.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.nickname
		FROM users u
		JOIN followers f ON u.id = f.followed_id
		WHERE f.follower_id = ? AND f.status = 'accepted'
	`, userID)
	if err != nil {
		http.Error(w, "Failed to get following", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	following := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Avatar, &user.Nickname); err != nil {
			http.Error(w, "Failed to scan following", http.StatusInternalServerError)
			return
		}
		following = append(following, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(following)
}
