package auth

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSessionToken creates a new session token.
func GenerateSessionToken() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// CreateSession creates a new session for a user.
func CreateSession(db *sql.DB, userID int64) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	stmt, err := db.Prepare("INSERT INTO sessions (user_id, token, expiry) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	expiry := time.Now().Add(24 * time.Hour)
	_, err = stmt.Exec(userID, token, expiry)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserFromSession retrieves a user ID from a session token.
func GetUserFromSession(db *sql.DB, token string) (int64, error) {
	var userID int64
	err := db.QueryRow("SELECT user_id FROM sessions WHERE token = ? AND expiry > ?", token, time.Now()).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteSession removes a session from the database.
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// DeleteUserSessions removes all sessions for a given user.
func DeleteUserSessions(db *sql.DB, userID int64) error {
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}
