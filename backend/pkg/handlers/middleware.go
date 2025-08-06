package handlers

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

func (env *Env) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionToken := cookie.Value
		var userID int64
		var expiresAt time.Time

		err = env.DB.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ?", sessionToken).Scan(&userID, &expiresAt)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if expiresAt.Before(time.Now()) {
			// Session has expired, delete it
			env.DB.Exec("DELETE FROM sessions WHERE token = ?", sessionToken)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
