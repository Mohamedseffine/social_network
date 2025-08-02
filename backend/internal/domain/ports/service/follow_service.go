package service

import "social_network/internal/domain/models"

type FollowService interface {
	CreateFollow(follow *models.Follow) error
	CreateFollowByUsername(follow *models.FollowByUsername, followerID int) error
	AcceptFollow(followerID, followingID, currentUserID int) error
	DeclineFollow(followerID, followingID, currentUserID int) error
	DeleteFollow(followerID, followingID, currentUserID int) error
	GetStatusFollow(followerID, followingID int) (string, error)
	GetFollowers(username string) ([]models.FollowerInfo, error)
	GetFollowing(username string) ([]models.FollowerInfo, error)
}
