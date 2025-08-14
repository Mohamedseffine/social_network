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
		SELECT id, user_id, content, image, privacy, created_at, first_name, last_name, avatar
		FROM (
			-- Regular posts from followed users and self
			SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at, u.first_name, u.last_name, u.avatar
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE (
				-- User's own posts
				p.user_id = ?
				-- All public posts from all users
				OR p.privacy = 'public'
				-- Almost private posts from followed users
				OR (p.user_id IN (SELECT followed_id FROM followers WHERE follower_id = ? AND status = 'accepted') AND p.privacy = 'almost private')
				-- Private posts user has access to
				OR (p.privacy = 'private' AND EXISTS (SELECT 1 FROM post_viewers pv WHERE pv.post_id = p.id AND pv.viewer_id = ?))
			)
			UNION ALL
			-- Group posts from user's groups
			SELECT gp.id, gp.user_id, gp.content, gp.image, 'group' as privacy, gp.created_at, u.first_name, u.last_name, u.avatar
			FROM group_posts gp
			JOIN users u ON gp.user_id = u.id
			WHERE gp.group_id IN (SELECT group_id FROM group_members WHERE user_id = ? AND status = 'accepted')
		) AS feed
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args := []interface{}{currentUserID, currentUserID, currentUserID, currentUserID, limit, offset}

	rows, err := app.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get feed posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var privacy, image, avatar sql.NullString
		if err := rows.Scan(&post.ID, &post.UserID, &post.Content, &image, &privacy, &post.CreatedAt, &post.AuthorFirstName, &post.AuthorLastName, &avatar); err != nil {
			http.Error(w, "Failed to scan feed post", http.StatusInternalServerError)
			return
		}
		if privacy.Valid {
			post.Privacy = privacy.String
		}
		if image.Valid {
			post.Image = image.String
		}
		if avatar.Valid && avatar.String != "" {
			post.AuthorAvatar = avatar.String
		} else {
			post.AuthorAvatar = "/uploads/default-avatar-icon-of-social-media-user-vector.jpg"
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
