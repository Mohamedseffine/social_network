package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/pkg/models"
	"strconv"
)

func (app *App) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 10
	offset := 0
	pageStr := r.URL.Query().Get("page")
	if page, err := strconv.Atoi(pageStr); err == nil && page > 1 {
		offset = (page - 1) * limit
	}

	query := `
		SELECT id, user_id, content, image, privacy, created_at, author_first_name, author_last_name, author_avatar
		FROM (
			-- Regular posts from followed users and self
			SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at, u.first_name AS author_first_name, u.last_name AS author_last_name, u.avatar AS author_avatar
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.group_id IS NULL AND (
				p.user_id = ?
				OR (p.user_id IN (SELECT followed_id FROM followers WHERE follower_id = ? AND status = 'accepted') AND p.privacy IN ('public', 'private'))
			)
			UNION ALL
			-- Group posts from user's groups
			SELECT gp.id, gp.user_id, gp.content, gp.image, 'group' as privacy, gp.created_at, u.first_name AS author_first_name, u.last_name AS author_last_name, u.avatar AS author_avatar
			FROM group_posts gp
			JOIN users u ON gp.user_id = u.id
			WHERE gp.group_id IN (SELECT group_id FROM group_members WHERE user_id = ? AND status = 'accepted')
		) AS feed
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args := []interface{}{currentUserID, currentUserID, currentUserID, limit, offset}

	rows, err := app.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get feed posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var privacy sql.NullString
		if err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Image, &privacy, &post.CreatedAt, &post.AuthorFirstName, &post.AuthorLastName, &post.AuthorAvatar); err != nil {
			http.Error(w, "Failed to scan feed post", http.StatusInternalServerError)
			return
		}
		if privacy.Valid {
			post.Privacy = privacy.String
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over feed posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
