package models

import "time"

type User struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	Password        string    `json:"-"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	DateOfBirth     string    `json:"date_of_birth"`
	Avatar          string    `json:"avatar,omitempty"`
	Nickname        string    `json:"nickname,omitempty"`
	AboutMe         string    `json:"about_me,omitempty"`
	ProfileIsPublic bool      `json:"profile_is_public"`
	CreatedAt       time.Time `json:"created_at"`
	FollowStatus    string    `json:"follow_status,omitempty"`
}
