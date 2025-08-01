package service

import (
	"net/http"

	"social_network/internal/domain/models"
)

// Create an interface for the websocket implementation:
type WebsocketSevice interface {
	CreateNewWebSocket(w http.ResponseWriter, r *http.Request) error
	GetAllUsersWithStatus(id, offset, limit int) ([]*models.ChatUser, error)
}
