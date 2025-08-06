package models

type Post struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	ImageURL  string `json:"image_url,omitempty"`
	Privacy   string `json:"privacy"`
	CreatedAt string `json:"created_at"`
}
