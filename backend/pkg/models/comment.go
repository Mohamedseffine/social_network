package models

import "time"

type Comment struct {
	ID              int64     `json:"id"`
	PostID          int64     `json:"post_id"`
	UserID          int64     `json:"user_id"`
	AuthorFirstName string    `json:"author_first_name"`
	AuthorLastName  string    `json:"author_last_name"`
	AuthorAvatar    string    `json:"author_avatar"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
}
