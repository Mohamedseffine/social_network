package models

import "time"

type Notification struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	IsRead     bool      `json:"is_read"`
	RelatedID  *int64    `json:"related_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
