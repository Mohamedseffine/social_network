package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (env *Env) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user ID of the person making the request from the context.
	loggedInUserID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the user ID of the person to follow from the URL.
	vars := mux.Vars(r)
	followUserIDStr := vars["id"]
	followUserID, err := strconv.ParseInt(followUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if loggedInUserID == followUserID {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	// Check the profile type of the user to be followed.
	var profileType string
	err = env.DB.QueryRow("SELECT profile_type FROM users WHERE id = ?", followUserID).Scan(&profileType)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	status := "pending"
	if profileType == "public" {
		status = "accepted"
	}

	// Insert the follow relationship into the database.
	_, err = env.DB.Exec(
		"INSERT INTO followers (follower_id, following_id, status) VALUES (?, ?, ?)",
		loggedInUserID, followUserID, status,
	)
	if err != nil {
		// Handle potential unique constraint violation
		http.Error(w, "Follow request already sent or user already followed", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (env *Env) HandleFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleFollowRequestHandler called")
	loggedInUserID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	requestIDStr := vars["id"]
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var action struct {
		Action string `json:"action"` // "accept" or "decline"
	}
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify that the logged-in user is the recipient of the follow request.
	var followingID int64
	err = env.DB.QueryRow("SELECT following_id FROM followers WHERE id = ?", requestID).Scan(&followingID)
	if err != nil {
		http.Error(w, "Follow request not found", http.StatusNotFound)
		return
	}

	if loggedInUserID != followingID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if action.Action == "accept" {
		_, err = env.DB.Exec("UPDATE followers SET status = 'accepted' WHERE id = ?", requestID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else if action.Action == "decline" {
		_, err = env.DB.Exec("DELETE FROM followers WHERE id = ?", requestID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}
