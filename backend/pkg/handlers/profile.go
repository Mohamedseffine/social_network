package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/models"
	"strconv"

	"github.com/gorilla/mux"
	"strings"
)

func (env *Env) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	loggedInUserID, _ := r.Context().Value(UserIDKey).(int64)

	vars := mux.Vars(r)
	profileUserID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	err = env.DB.QueryRow(
		"SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type FROM users WHERE id = ?",
		profileUserID,
	).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.Avatar, &user.Nickname, &user.AboutMe, &user.ProfileType,
	)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// If the profile is private, check for permission
	if user.ProfileType == "private" {
		isOwner := loggedInUserID == user.ID
		isFollower := false
		if loggedInUserID != 0 {
			var status string
			err := env.DB.QueryRow(
				"SELECT status FROM followers WHERE follower_id = ? AND following_id = ?",
				loggedInUserID, user.ID,
			).Scan(&status)
			if err == nil && status == "accepted" {
				isFollower = true
			}
		}

		if !isOwner && !isFollower {
			http.Error(w, "This profile is private", http.StatusForbidden)
			return
		}
	}

	// Do not expose password
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (env *Env) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	loggedInUserID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build the UPDATE query dynamically
	query := "UPDATE users SET"
	args := []interface{}{}

	if updatedUser.FirstName != "" {
		query += " first_name = ?,"
		args = append(args, updatedUser.FirstName)
	}
	if updatedUser.LastName != "" {
		query += " last_name = ?,"
		args = append(args, updatedUser.LastName)
	}
	if updatedUser.Nickname != "" {
		query += " nickname = ?,"
		args = append(args, updatedUser.Nickname)
	}
	if updatedUser.AboutMe != "" {
		query += " about_me = ?,"
		args = append(args, updatedUser.AboutMe)
	}
	if updatedUser.Avatar != "" {
		query += " avatar = ?,"
		args = append(args, updatedUser.Avatar)
	}
	if updatedUser.ProfileType != "" {
		if updatedUser.ProfileType != "public" && updatedUser.ProfileType != "private" {
			http.Error(w, "Invalid profile type", http.StatusBadRequest)
			return
		}
		query += " profile_type = ?,"
		args = append(args, updatedUser.ProfileType)
	}

	// Remove the trailing comma
	query = strings.TrimSuffix(query, ",")

	query += " WHERE id = ?"
	args = append(args, loggedInUserID)

	// Only execute if there's something to update
	if len(args) > 1 {
		_, err := env.DB.Exec(query, args...)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
