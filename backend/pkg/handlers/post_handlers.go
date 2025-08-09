package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"

	"github.com/gorilla/mux"
)

type CreatePostRequest struct {
	Content string `json:"content"`
	Image   string `json:"image"`
	Privacy string `json:"privacy"`
}

func (app *App) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("INSERT INTO posts (user_id, content, image, privacy) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, req.Content, req.Image, req.Privacy)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Post created successfully"))
}

func (app *App) GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profileUserIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}

	profileUserID, err := strconv.ParseInt(profileUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	requestingUserID := ForContext(r.Context())

	isOwner := requestingUserID == profileUserID
	isFollowing, err := app.isFollowing(requestingUserID, profileUserID)
	if err != nil {
		http.Error(w, "Failed to check follow status", http.StatusInternalServerError)
		return
	}

	query := "SELECT id, user_id, content, image, privacy, created_at FROM posts WHERE user_id = ?"
	args := []interface{}{profileUserID}

	if !isOwner {
		if isFollowing {
			query += " AND privacy IN ('public', 'private')"
		} else {
			query += " AND privacy = 'public'"
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := app.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt); err != nil {
			http.Error(w, "Failed to scan post", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "Failed to encode posts", http.StatusInternalServerError)
	}
}
