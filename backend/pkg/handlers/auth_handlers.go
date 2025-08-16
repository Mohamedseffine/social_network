package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/pkg/auth"
	"social-network/backend/pkg/models"
	"time"
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Avatar      string `json:"avatar"`
	Nickname    string `json:"nickname"`
	AboutMe     string `json:"about_me"`
}

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	now := time.Now() 
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age-- 
	}

	if age < 18 {
		http.Error(w, "You must be at least 18 years old to register", http.StatusBadRequest)
		return
	}

	if req.Avatar == "" {
		req.Avatar = "/uploads/default-avatar-icon-of-social-media-user-vector.jpg"
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	stmt, err := app.DB.Prepare(`
		INSERT INTO users (email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(req.Email, hashedPassword, req.FirstName, req.LastName, req.DateOfBirth, req.Avatar, req.Nickname, req.AboutMe)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}



func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := app.DB.QueryRow("SELECT id, password FROM users WHERE email = ?", credentials.Email).Scan(&user.ID, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPasswordHash(credentials.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Delete any existing sessions for this user
	if err := auth.DeleteUserSessions(app.DB, user.ID); err != nil {
		http.Error(w, "Failed to clear existing sessions", http.StatusInternalServerError)
		return
	}

	token, err := auth.CreateSession(app.DB, user.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	// We need to fetch the full user object to return it
	var avatar, nickname, aboutMe sql.NullString
	err = app.DB.QueryRow("SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_is_public, created_at FROM users WHERE id = ?", user.ID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DateOfBirth, &avatar, &nickname, &aboutMe, &user.ProfileIsPublic, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to get full user details", http.StatusInternalServerError)
		return
	}
	if avatar.Valid && avatar.String != "" {
		user.Avatar = avatar.String
	} else {
		user.Avatar = "/uploads/default-avatar-icon-of-social-media-user-vector.jpg"
	}
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if aboutMe.Valid {
		user.AboutMe = aboutMe.String
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (app *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Not logged in", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Invalid cookie", http.StatusBadRequest)
		return
	}

	token := cookie.Value
	if err := auth.DeleteSession(app.DB, token); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}

func (app *App) GetSessionUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var user models.User
	var avatar, nickname, aboutMe sql.NullString
	err := app.DB.QueryRow("SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_is_public, created_at FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.DateOfBirth, &avatar, &nickname, &aboutMe, &user.ProfileIsPublic, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if avatar.Valid && avatar.String != "" {
		user.Avatar = avatar.String
	} else {
		user.Avatar = "/uploads/default-avatar-icon-of-social-media-user-vector.jpg"
	}
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if aboutMe.Valid {
		user.AboutMe = aboutMe.String
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
