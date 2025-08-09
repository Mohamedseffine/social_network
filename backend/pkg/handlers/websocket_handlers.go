package handlers

import (
	"net/http"

	"social-network/backend/pkg/websockets"
)

func (app *App) ServeWs(hub *websockets.Hub, w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	websockets.ServeWs(hub, w, r, userID)
}
