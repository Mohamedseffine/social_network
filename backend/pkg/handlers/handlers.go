package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"

	"github.com/gorilla/mux"
)

type App struct {
	DB *sql.DB
}

func (app *App) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}

	profileUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	requestingUserID := ForContext(r.Context())

	var user models.User
	err = app.DB.QueryRow("SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_is_public, created_at FROM users WHERE id = ?", profileUserID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DateOfBirth, &user.Avatar, &user.Nickname, &user.AboutMe, &user.ProfileIsPublic, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Privacy check
	if !user.ProfileIsPublic {
		isOwner := requestingUserID == profileUserID
		isFollowing, _ := app.isFollowing(requestingUserID, profileUserID)
		if !isOwner && !isFollowing {
			// Return a limited profile for private users to non-followers
			limitedUser := models.User{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Nickname:  user.Nickname,
				Avatar:    user.Avatar,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(limitedUser)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)
	}
}

func (app *App) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, email, first_name, last_name, avatar, nickname FROM users")
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Avatar, &user.Nickname); err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode users", http.StatusInternalServerError)
	}
}

// isFollowing checks if a user is following another user.
func (app *App) isFollowing(followerID, followedID int64) (bool, error) {
	if followerID == 0 {
		return false, nil
	}

	var status string
	err := app.DB.QueryRow("SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?", followerID, followedID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return status == "accepted", nil
}
