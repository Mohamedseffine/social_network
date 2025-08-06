package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/models"
)

func (env *Env) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	loggedInUserID, _ := r.Context().Value(UserIDKey).(int64)

	var newPost models.Post
	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate privacy setting
	if newPost.Privacy != "public" && newPost.Privacy != "private" && newPost.Privacy != "almost_private" {
		http.Error(w, "Invalid privacy setting", http.StatusBadRequest)
		return
	}

	newPost.UserID = loggedInUserID

	result, err := env.DB.Exec(
		"INSERT INTO posts (user_id, content, image_url, privacy) VALUES (?, ?, ?, ?)",
		newPost.UserID, newPost.Content, newPost.ImageURL, newPost.Privacy,
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()
	newPost.ID = postID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (env *Env) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	loggedInUserID, _ := r.Context().Value(UserIDKey).(int64)

	// This query gets:
	// 1. All public posts.
	// 2. All "almost private" posts from users the logged-in user follows.
	rows, err := env.DB.Query(`
		SELECT p.id, p.user_id, p.content, p.image_url, p.privacy, p.created_at
		FROM posts p
		LEFT JOIN followers f ON p.user_id = f.following_id AND f.follower_id = ?
		WHERE
			p.privacy = 'public' OR
			(p.privacy = 'almost_private' AND f.status = 'accepted') OR
			p.user_id = ?
		ORDER BY p.created_at DESC
	`, loggedInUserID, loggedInUserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.Privacy, &post.CreatedAt); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
