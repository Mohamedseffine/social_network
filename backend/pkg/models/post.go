package models

import "time"

type Post struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	AuthorFirstName string    `json:"author_first_name"`
	AuthorLastName  string    `json:"author_last_name"`
	AuthorAvatar    string    `json:"author_avatar"`
	Content         string    `json:"content"`
	Image           string    `json:"image,omitempty"`
	Privacy         string    `json:"privacy,omitempty"` // Privacy may not apply to group posts
	CreatedAt       time.Time `json:"created_at"`
}
