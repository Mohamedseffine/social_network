package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social-network/backend/pkg/models"
	"social-network/backend/pkg/router"
)

type CreateGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (app *App) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := app.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to prepare group statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, req.Title, req.Description)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to get group ID", http.StatusInternalServerError)
		return
	}

	memberStmt, err := tx.Prepare("INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to prepare member statement", http.StatusInternalServerError)
		return
	}
	defer memberStmt.Close()

	_, err = memberStmt.Exec(groupID, userID, "accepted")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to add creator to group", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Group created successfully"))
}

func (app *App) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, creator_id, title, description, created_at FROM groups")
	if err != nil {
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.CreatedAt); err != nil {
			http.Error(w, "Failed to scan group", http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		http.Error(w, "Failed to encode groups", http.StatusInternalServerError)
	}
}

func (app *App) GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var group models.Group
	err = app.DB.QueryRow("SELECT id, creator_id, title, description, created_at FROM groups WHERE id = ?", groupID).Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get group", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(group); err != nil {
		http.Error(w, "Failed to encode group", http.StatusInternalServerError)
	}
}

func (app *App) GetGroupMembershipStatusHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, currentUserID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			status = "not_member"
		} else {
			http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]string{"status": status}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (app *App) JoinGroupHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the user is already a member or has a pending request
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}
	if err == nil {
		http.Error(w, "You are already a member or have a pending request for this group", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(groupID, userID, "pending")
	if err != nil {
		http.Error(w, "Failed to join group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Join request sent"))

	// Create notification for the group owner
	joinRequestID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID for group join request: %v", err)
	} else {
		var creatorID int64
		err := app.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
		if err != nil {
			log.Printf("Failed to get group creator ID for notification: %v", err)
		} else {
			var group models.Group
			err := app.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&group.Title)
			if err != nil {
				log.Printf("Failed to get group info for notification: %v", err)
			} else {
				message := fmt.Sprintf("wants to join your group '%s'.", group.Title)
				app.createNotification(creatorID, "group_join_request", message, userID, joinRequestID)
			}
		}
	}
}

func (app *App) AcceptGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	memberIDStr := router.ForContext(r.Context(), "id")
	if memberIDStr == "" {
		http.Error(w, "Member ID is missing", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Get the group id from the member id
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM group_members WHERE id = ?", memberID).Scan(&groupID)
	if err != nil {
		http.Error(w, "Failed to get group ID from member ID", http.StatusInternalServerError)
		return
	}

	// Check if the current user is the creator of the group
	var creatorID int64
	err = app.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err != nil {
		http.Error(w, "Failed to get group creator ID", http.StatusInternalServerError)
		return
	}

	if currentUserID != creatorID {
		http.Error(w, "Only the group creator can accept requests", http.StatusForbidden)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE group_members SET status = 'accepted' WHERE id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(memberID)
	if err != nil {
		http.Error(w, "Failed to accept join request", http.StatusInternalServerError)
		return
	}

	// Delete the notification
	_, err = app.DB.Exec("DELETE FROM notifications WHERE type = 'group_join_request' AND related_id = ?", memberID)
	if err != nil {
		log.Printf("Failed to delete group join notification: %v", err)
	}

	// Notify the user that their request was accepted
	var requestingUserID int64
	var groupName string
	err = app.DB.QueryRow("SELECT user_id FROM group_members WHERE id = ?", memberID).Scan(&requestingUserID)
	if err != nil {
		log.Printf("Failed to get user ID for notification: %v", err)
	} else {
		err = app.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&groupName)
		if err != nil {
			log.Printf("Failed to get group name for notification: %v", err)
		} else {
			app.createNotification(requestingUserID, "group_request_accepted", fmt.Sprintf("Your request to join '%s' has been accepted.", groupName), currentUserID, groupID)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Join request accepted"))
}

func (app *App) DeclineGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	memberIDStr := router.ForContext(r.Context(), "id")
	if memberIDStr == "" {
		http.Error(w, "Member ID is missing", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Get the group id from the member id to check for creator
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM group_members WHERE id = ?", memberID).Scan(&groupID)
	if err != nil {
		http.Error(w, "Failed to get group ID from member ID", http.StatusInternalServerError)
		return
	}

	// Check if the current user is the creator of the group
	var creatorID int64
	err = app.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err != nil {
		http.Error(w, "Failed to get group creator ID", http.StatusInternalServerError)
		return
	}

	if currentUserID != creatorID {
		http.Error(w, "Only the group creator can decline requests", http.StatusForbidden)
		return
	}

	// Get the user ID who made the request BEFORE deleting the record
	var requestingUserID int64
	err = app.DB.QueryRow("SELECT user_id FROM group_members WHERE id = ? AND status = 'pending'", memberID).Scan(&requestingUserID)
	if err != nil {
		http.Error(w, "Failed to find pending join request", http.StatusInternalServerError)
		return
	}

	// Delete the pending request
	stmt, err := app.DB.Prepare("DELETE FROM group_members WHERE id = ?")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(memberID)
	if err != nil {
		http.Error(w, "Failed to decline join request", http.StatusInternalServerError)
		return
	}

	// Delete the notification
	_, err = app.DB.Exec("DELETE FROM notifications WHERE type = 'group_join_request' AND related_id = ?", memberID)
	if err != nil {
		log.Printf("Failed to delete group join notification: %v", err)
	}

	// Notify the user that their request was declined
	var groupName string
	err = app.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&groupName)
	if err != nil {
		log.Printf("Failed to get group name for notification: %v", err)
	} else {
		app.createNotification(requestingUserID, "group_request_declined", fmt.Sprintf("Your request to join '%s' has been declined.", groupName), currentUserID, groupID)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Join request declined"))
}

type InviteToGroupRequest struct {
	UserID int64 `json:"user_id"`
}

func (app *App) InviteToGroupHandler(w http.ResponseWriter, r *http.Request) {
	invitingUserID := ForContext(r.Context())
	if invitingUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req InviteToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the inviting user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, invitingUserID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can invite others", http.StatusForbidden)
		return
	}

	// Check if the invited user is already a member or has a pending request/invitation
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, req.UserID).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}
	if err == nil {
		http.Error(w, "This user is already a member or has a pending request/invitation", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(groupID, req.UserID, "invited")
	if err != nil {
		http.Error(w, "Failed to invite user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User invited successfully"))

	// Create notification
	inviteID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID for group invite: %v", err)
	} else {
		var group models.Group
		err := app.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&group.Title)
		if err != nil {
			log.Printf("Failed to get group info for notification: %v", err)
		} else {
			message := fmt.Sprintf("invited you to join the group '%s'.", group.Title)
			app.createNotification(req.UserID, "group_invite", message, invitingUserID, inviteID)
		}
	}
}

func (app *App) AcceptGroupInviteHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	inviteIDStr := router.ForContext(r.Context(), "id")
	if inviteIDStr == "" {
		http.Error(w, "Invite ID is missing", http.StatusBadRequest)
		return
	}

	inviteID, err := strconv.ParseInt(inviteIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invite ID", http.StatusBadRequest)
		return
	}

	// Check if the current user is the one who was invited
	var invitedUserID int64
	err = app.DB.QueryRow("SELECT user_id FROM group_members WHERE id = ?", inviteID).Scan(&invitedUserID)
	if err != nil {
		http.Error(w, "Failed to get invited user ID", http.StatusInternalServerError)
		return
	}

	if currentUserID != invitedUserID {
		http.Error(w, "You are not authorized to accept this invitation", http.StatusForbidden)
		return
	}

	stmt, err := app.DB.Prepare("UPDATE group_members SET status = 'accepted' WHERE id = ? AND status = 'invited'")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(inviteID)
	if err != nil {
		http.Error(w, "Failed to accept invitation", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Invitation not found or already accepted", http.StatusNotFound)
		return
	}

	// Delete the notification
	_, err = app.DB.Exec("DELETE FROM notifications WHERE type = 'group_invite' AND related_id = ?", inviteID)
	if err != nil {
		log.Printf("Failed to delete group invite notification: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Invitation accepted"))
}

func (app *App) DeclineGroupInviteHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	inviteIDStr := router.ForContext(r.Context(), "id")
	if inviteIDStr == "" {
		http.Error(w, "Invite ID is missing", http.StatusBadRequest)
		return
	}

	inviteID, err := strconv.ParseInt(inviteIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invite ID", http.StatusBadRequest)
		return
	}

	// Check if the current user is the one who was invited
	var invitedUserID int64
	err = app.DB.QueryRow("SELECT user_id FROM group_members WHERE id = ?", inviteID).Scan(&invitedUserID)
	if err != nil {
		http.Error(w, "Failed to get invited user ID", http.StatusInternalServerError)
		return
	}

	if currentUserID != invitedUserID {
		http.Error(w, "You are not authorized to decline this invitation", http.StatusForbidden)
		return
	}

	stmt, err := app.DB.Prepare("DELETE FROM group_members WHERE id = ? AND status = 'invited'")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(inviteID)
	if err != nil {
		http.Error(w, "Failed to decline invitation", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Invitation not found or already handled", http.StatusNotFound)
		return
	}

	// Delete the notification
	_, err = app.DB.Exec("DELETE FROM notifications WHERE type = 'group_invite' AND related_id = ?", inviteID)
	if err != nil {
		log.Printf("Failed to delete group invite notification: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Invitation declined"))
}

type CreateGroupPostRequest struct {
	Content string `json:"content"`
	Image   string `json:"image"`
}

func (app *App) CreateGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can post in the group", http.StatusForbidden)
		return
	}

	var req CreateGroupPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Println(req.Image)

	stmt, err := app.DB.Prepare("INSERT INTO group_posts (group_id, user_id, content, image) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(groupID, userID, req.Content, req.Image)
	if err != nil {
		http.Error(w, "Failed to create group post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Group post created successfully"))
}

func (app *App) GetGroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	groupIDStr := router.ForContext(r.Context(), "id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is missing", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to check membership status", http.StatusInternalServerError)
		return
	}

	if status != "accepted" {
		http.Error(w, "Only accepted members can view group posts", http.StatusForbidden)
		return
	}

	rows, err := app.DB.Query("SELECT id, group_id, user_id, content, image, created_at FROM group_posts WHERE group_id = ? ORDER BY created_at DESC", groupID)
	if err != nil {
		http.Error(w, "Failed to get group posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.GroupPost
	for rows.Next() {
		var post models.GroupPost
		var image sql.NullString
		if err := rows.Scan(&post.ID, &post.GroupID, &post.UserID, &post.Content, &image, &post.CreatedAt); err != nil {
			http.Error(w, "Failed to scan group post", http.StatusInternalServerError)
			return
		}
		if image.Valid {
			post.Image = image.String
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over group posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "Failed to encode group posts", http.StatusInternalServerError)
	}
}

type CreateGroupPostCommentRequest struct {
	Content  string `json:"content"`
	ImageUrl string `json:"image"`
}

func (app *App) CreateGroupPostCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	postIDStr := router.ForContext(r.Context(), "id")
	if postIDStr == "" {
		http.Error(w, "Post ID is missing", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get group_id from post_id to check membership
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM group_posts WHERE id = ?", postID).Scan(&groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get group ID from post", http.StatusInternalServerError)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil || status != "accepted" {
		http.Error(w, "Only accepted members of the group can comment", http.StatusForbidden)
		return
	}

	var req CreateGroupPostCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	stmt, err := app.DB.Prepare("INSERT INTO group_post_comments (post_id, user_id, content, image_url) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(postID, userID, req.Content, req.ImageUrl)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Comment created successfully"))
}

func (app *App) GetGroupPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	postIDStr := router.ForContext(r.Context(), "id")
	if postIDStr == "" {
		http.Error(w, "Post ID is missing", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get group_id from post_id to check membership
	var groupID int64
	err = app.DB.QueryRow("SELECT group_id FROM group_posts WHERE id = ?", postID).Scan(&groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get group ID from post", http.StatusInternalServerError)
		return
	}

	// Check if the user is a member of the group
	var status string
	err = app.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&status)
	if err != nil || status != "accepted" {
		http.Error(w, "Only accepted members of the group can view comments", http.StatusForbidden)
		return
	}

	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, c.image_url, u.first_name, u.last_name, u.avatar
		FROM group_post_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`
	rows, err := app.DB.Query(query, postID)
	if err != nil {
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.GroupPostComment
	for rows.Next() {
		var comment models.GroupPostComment
		var avatar, image sql.NullString
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &image, &comment.AuthorFirstName, &comment.AuthorLastName, &avatar); err != nil {
			http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
			return
		}
		if avatar.Valid && avatar.String != "" {
			comment.AuthorAvatar = avatar.String
		} else {
			comment.AuthorAvatar = "/uploads/default-avatar-icon-of-social-media-user-vector.jpg"
		}
		if image.Valid && image.String != "" {
			comment.ImageUrl = image.String
		} else {
			comment.ImageUrl = ""
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		http.Error(w, "Failed to encode comments", http.StatusInternalServerError)
	}
}

func (app *App) SearchGroupsHandler(w http.ResponseWriter, r *http.Request) {
	userID := ForContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	queryValues := r.URL.Query()
	searchQuery := queryValues.Get("q")
	if searchQuery == "" {
		http.Error(w, "the search value is empty", http.StatusBadRequest)
		return
	}
	stm, err := app.DB.Prepare(`SELECT id, title, description FROM groups WHERE title LIKE ? OR description LIKE ? `)
	if err != nil {
		log.Println(err, "1")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := stm.Query("%"+searchQuery+"%", "%"+searchQuery+"%")
	if err != nil {
		log.Println(err, "2")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.ID, &group.Title, &group.Description)
		if err != nil {
			log.Println(err, "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, group)
	}
	log.Println("res", result)
	err = json.NewEncoder(w).Encode(map[string]any{
		"groups": result,
	})
	if err != nil {
		log.Println(err, "4")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
