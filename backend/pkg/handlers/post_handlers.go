package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"

	"github.com/gorilla/mux"
)

type CreatePostRequest struct {
	Content   string  `json:"content"`
	Image     string  `json:"image"`
	Privacy   string  `json:"privacy"`
	ViewerIDs []int64 `json:"viewer_ids,omitempty"`
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

	tx, err := app.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Rollback on any error

	// Insert post
	stmt, err := tx.Prepare("INSERT INTO posts (user_id, content, image, privacy) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare post statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, req.Content, req.Image, req.Privacy)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get post ID", http.StatusInternalServerError)
		return
	}

	// If post is private, validate viewer IDs and insert into post_viewers
	if req.Privacy == "private" {
		// Server-side validation: ensure all viewerIDs are actual followers
		for _, viewerID := range req.ViewerIDs {
			isFollower, err := app.isFollowing(viewerID, userID)
			if err != nil {
				http.Error(w, "Failed to validate follower status", http.StatusInternalServerError)
				return
			}
			if !isFollower {
				http.Error(w, "Invalid viewer ID: user is not a follower.", http.StatusBadRequest)
				return
			}
		}

		viewerStmt, err := tx.Prepare("INSERT INTO post_viewers (post_id, viewer_id) VALUES (?, ?)")
		if err != nil {
			http.Error(w, "Failed to prepare viewer statement", http.StatusInternalServerError)
			return
		}
		defer viewerStmt.Close()

		for _, viewerID := range req.ViewerIDs {
			_, err := viewerStmt.Exec(postID, viewerID)
			if err != nil {
				http.Error(w, "Failed to add viewer to post", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
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

	query := `
		SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
		u.first_name, u.last_name, u.avatar
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?`
	args := []interface{}{profileUserID}

	if !isOwner {
		privacyClause := " AND (p.privacy = 'public'"
		if isFollowing {
			privacyClause += " OR p.privacy = 'almost private'"
		}
		// Check for private posts the user has access to
		privacyClause += " OR (p.privacy = 'private' AND EXISTS (SELECT 1 FROM post_viewers WHERE post_id = p.id AND viewer_id = ?)))"
		query += privacyClause
		args = append(args, requestingUserID)
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := app.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt, &post.AuthorFirstName, &post.AuthorLastName, &post.AuthorAvatar); err != nil {
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
