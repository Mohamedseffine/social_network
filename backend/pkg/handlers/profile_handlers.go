package handlers

import (
	"encoding/json"
	"net/http"
)

type UpdateProfilePrivacyRequest struct {
	IsPublic bool `json:"is_public"`
}

func (app *App) UpdateProfilePrivacyHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req UpdateProfilePrivacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE users SET profile_is_public = ? WHERE id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(req.IsPublic, userID)
	if err != nil {
		http.Error(w, "Failed to update profile privacy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profile privacy updated successfully"))
}
