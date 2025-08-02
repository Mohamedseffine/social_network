package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"social_network/internal/adapters/http/utils"
	"social_network/internal/domain/models"
	"social_network/internal/domain/ports/service"
)

type FollowHandler struct {
	followService  service.FollowService
	sessionService service.SessionService
}

func NewFollowHandler(followSvc service.FollowService, sessionSvc service.SessionService) *FollowHandler {
	return &FollowHandler{
		followService:  followSvc,
		sessionService: sessionSvc,
	}
}

	func (h *FollowHandler) CreateFollow(w http.ResponseWriter, r *http.Request) {
		
		if r.Method != http.MethodPost {
			utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
			return
		}

		var payload struct {
			FollowingID int `json:"target_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
			return
		}

		// Get the current user's ID from the session
		followerID, err := utils.GetCurrentUserID(r, h.sessionService)
		if err != nil {
			fmt.Println("Error getting current user ID:", err)
			utils.ResponseJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
			return
		}

		follow := &models.Follow{
			FollowerID:  followerID,
			FollowingID: payload.FollowingID,
		}

		if err := h.followService.CreateFollow(follow); err != nil {
			utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		utils.ResponseJSON(w, http.StatusCreated, map[string]any{
			"success": true,
			"message": "Follow request created successfully.",
		})
	}

func (h *FollowHandler) CreateFollowByUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	var payload models.FollowByUsername

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		return
	}

	// Get the current user's ID from the session
	followerID, err := utils.GetCurrentUserID(r, h.sessionService)
	if err != nil {
		utils.ResponseJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
		return
	}

	if err := h.followService.CreateFollowByUsername(&payload, followerID); err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Follow request created successfully.",
	})
}


func (h *FollowHandler) AcceptFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	var payload struct {
		FollowerID  int `json:"follower_id"`
		FollowingID int `json:"following_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		return
	}

	currentUserID, err := utils.GetCurrentUserID(r, h.sessionService)
	if err != nil {
		utils.ResponseJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
		return
	}

	if err := h.followService.AcceptFollow(payload.FollowerID, payload.FollowingID, currentUserID); err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Follow request accepted successfully.",
	})
}

func (h *FollowHandler) DeclineFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	var payload struct {
		FollowerID  int `json:"follower_id"`
		FollowingID int `json:"following_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		return
	}

	currentUserID, err := utils.GetCurrentUserID(r, h.sessionService)
	if err != nil {
		utils.ResponseJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
		return
	}

	if err := h.followService.DeclineFollow(payload.FollowerID, payload.FollowingID, currentUserID); err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Follow request declined successfully.",
	})
}

func (h *FollowHandler) DeleteFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	var payload struct {
		FollowerID  int `json:"follower_id"`
		FollowingID int `json:"following_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		return
	}

	currentUserID, err := utils.GetCurrentUserID(r, h.sessionService)
	if err != nil {
		utils.ResponseJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized"})
		return
	}

	if err := h.followService.DeleteFollow(payload.FollowerID, payload.FollowingID, currentUserID); err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Follow relationship deleted successfully.",
	})
}

func (h *FollowHandler) GetStatusFollow(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ffffffffffffffffffff")
	if r.Method != http.MethodGet {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}



	followingIDStr := r.URL.Query().Get("target_id")


	followerID, err := utils.GetCurrentUserID(r, h.sessionService)
	if err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid follower_id"})
		return
	}
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid following_id"})
		return
	}

	if followerID == 0 || followingID == 0 {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Follower and following IDs must be provided"})
		return
	}

	status, err := h.followService.GetStatusFollow(followerID, followingID)
	if err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, map[string]any{
		"status":  status,
		"success": true,
	})
}

func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetFollowers called")
	if r.Method != http.MethodGet {
		fmt.Println("Method not allowed")
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		fmt.Println("Invalid username")
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid username"})
		return
	}

	followers, err := h.followService.GetFollowers(username)
	if err != nil {
		fmt.Println("Error fetching followers:", err)
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, followers)
}

func (h *FollowHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ResponseJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Method not allowed"})
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		utils.ResponseJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid username"})
		return
	}

	following, err := h.followService.GetFollowing(username)
	if err != nil {
		utils.ResponseJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	utils.ResponseJSON(w, http.StatusOK, following)
}

//  Get the following status:
func (h *FollowHandler) GetFollowingStatus(w http.ResponseWriter, r *http.Request) {
    targetIDStr := r.URL.Query().Get("target_id")
    if targetIDStr == "" {
        http.Error(w, "missing target_id", http.StatusBadRequest)
        return
    }

    targetID, err := strconv.Atoi(targetIDStr)
    if err != nil {
        http.Error(w, "invalid target_id", http.StatusBadRequest)
        return
    }

	fmt.Println("Target id: ", map[string]int{"status": targetID})
    // Now you can use targetID...
	utils.ResponseJSON(w, http.StatusOK, map[string]string{"status": "pending"})
}

