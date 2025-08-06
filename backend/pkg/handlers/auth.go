package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"social-network/pkg/models"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Env struct {
	DB *sql.DB
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (env *Env) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LogoutHandler called")
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, there is no session to log out of.
			http.Error(w, "Not logged in", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value

	// Delete the session from the database
	_, err = env.DB.Exec("DELETE FROM sessions WHERE token = ?", sessionToken)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Expire the cookie on the client
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to a past time
		MaxAge:   -1,                         // This will delete the cookie
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (env *Env) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := env.DB.QueryRow("SELECT id, password FROM users WHERE email = ?", credentials.Email).Scan(&user.ID, &user.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = env.DB.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)", user.ID, token, expiresAt)
	if err != nil {
		log.Printf("Error inserting session into database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged in successfully"})
}

func (env *Env) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if user.Email == "" || user.Password == "" || user.FirstName == "" || user.LastName == "" || user.DateOfBirth == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store user in the database
	stmt, err := env.DB.Prepare(`
		INSERT INTO users (email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Email, string(hashedPassword), user.FirstName, user.LastName, user.DateOfBirth, user.Avatar, user.Nickname, user.AboutMe)
	if err != nil {
		// Check if the error is a unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Email already exists", http.StatusConflict)
		} else {
			// For any other error, return a generic internal server error
			log.Printf("Error inserting user into database: %v", err) // Log the actual error
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
