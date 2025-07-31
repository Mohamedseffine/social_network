package service

import "social_network/internal/domain/models"

type MessageService interface {
	GetChatHistoryService(id int, sessionValue string, offset int, limit int) ([]*models.PrivateMessage, error)
	MarkMessageAsRead(fromID, userId int) error
}
