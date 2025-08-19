package main

import (
	"log"
	"net/http"
	"social-network/backend/pkg/db/sqlite"
	"social-network/backend/pkg/handlers"
	"social-network/backend/pkg/router"
	"social-network/backend/pkg/websockets"
)

func main() {
	db := sqlite.InitDB("social-network.db")
	defer db.Close()

	hub := websockets.NewHub(db)
	go hub.Run()
	app := &handlers.App{DB: db, Hub: hub}

	r := router.NewRouter(db)

	// Public routes
	r.Handle("POST", "/api/register", app.RegisterHandler)
	r.Handle("POST", "/api/login", app.LoginHandler)
	r.Handle("POST", "/api/upload", app.UploadImageHandler)
	r.Handle("GET", "/uploads/{path}", app.GetImageHandler)

	// Authenticated routes
	authMiddleware := app.Authenticate
	r.Handle("POST", "/api/logout", app.LogoutHandler, authMiddleware)
	r.Handle("POST", "/api/users/{id}/follow", app.FollowUserHandler, authMiddleware)
	r.Handle("POST", "/api/users/{id}/unfollow", app.UnfollowUserHandler, authMiddleware)
	r.Handle("POST", "/api/requests/{id}/accept", app.AcceptFollowRequestHandler, authMiddleware)
	r.Handle("POST", "/api/requests/{id}/decline", app.DeclineFollowRequestHandler, authMiddleware)
	r.Handle("GET", "/api/users", app.GetAllUsersHandler, authMiddleware)
	r.Handle("GET", "/api/users/{id}", app.GetUserHandler, authMiddleware)
	r.Handle("GET", "/api/users/online", app.GetOnlineUsersHandler, authMiddleware)
	r.Handle("GET", "/api/users/{id}/followers", app.GetFollowersHandler, authMiddleware)
	r.Handle("GET", "/api/users/{id}/following", app.GetFollowingHandler, authMiddleware)
	r.Handle("GET", "/api/users/{id}/posts", app.GetUserPostsHandler, authMiddleware)
	r.Handle("POST", "/api/posts", app.CreatePostHandler, authMiddleware)
	r.Handle("POST", "/api/posts/{id}/comments", app.CreateCommentHandler, authMiddleware)
	r.Handle("GET", "/api/posts/{id}/comments", app.GetCommentsHandler, authMiddleware)
	r.Handle("POST", "/api/groups", app.CreateGroupHandler, authMiddleware)
	r.Handle("GET", "/api/groups", app.GetGroupsHandler, authMiddleware)
	r.Handle("GET", "/api/groups/{id}", app.GetGroupHandler, authMiddleware)
	r.Handle("GET", "/api/groups/{id}/membership", app.GetGroupMembershipStatusHandler, authMiddleware)
	r.Handle("POST", "/api/groups/{id}/join", app.JoinGroupHandler, authMiddleware)
	r.Handle("POST", "/api/groups/requests/{id}/accept", app.AcceptGroupRequestHandler, authMiddleware)
	r.Handle("POST", "/api/groups/requests/{id}/decline", app.DeclineGroupRequestHandler, authMiddleware)
	r.Handle("POST", "/api/groups/{id}/invite", app.InviteToGroupHandler, authMiddleware)
	r.Handle("POST", "/api/groups/invites/{id}/accept", app.AcceptGroupInviteHandler, authMiddleware)
	r.Handle("POST", "/api/groups/invites/{id}/decline", app.DeclineGroupInviteHandler, authMiddleware)
	r.Handle("POST", "/api/groups/{id}/posts", app.CreateGroupPostHandler, authMiddleware)
	r.Handle("GET", "/api/groups/{id}/posts", app.GetGroupPostsHandler, authMiddleware)
	r.Handle("POST", "/api/group-posts/{id}/comments", app.CreateGroupPostCommentHandler, authMiddleware)
	r.Handle("GET", "/api/group-posts/{id}/comments", app.GetGroupPostCommentsHandler, authMiddleware)
	r.Handle("GET", "/api/search_groups", app.SearchGroupsHandler, authMiddleware)
	r.Handle("POST", "/api/groups/{id}/events", app.CreateEventHandler, authMiddleware)
	r.Handle("GET", "/api/groups/{id}/events", app.GetGroupEventsHandler, authMiddleware)
	r.Handle("POST", "/api/events/{id}/respond", app.RespondToEventHandler, authMiddleware)
	r.Handle("GET", "/api/events/{id}/attendees", app.GetEventAttendeesHandler, authMiddleware)
	r.Handle("GET", "/api/notifications", app.GetNotificationsHandler, authMiddleware)
	r.Handle("GET", "/api/notifications/unread-count", app.GetUnreadNotificationCountHandler, authMiddleware)
	r.Handle("POST", "/api/notifications/read-all", app.MarkAllNotificationsAsReadHandler, authMiddleware)
	r.Handle("POST", "/api/notifications/{id}/read", app.MarkNotificationAsReadHandler, authMiddleware)
	r.Handle("DELETE", "/api/notifications/{id}", app.DeleteNotificationHandler, authMiddleware)
	r.Handle("GET", "/api/ws", func(w http.ResponseWriter, r *http.Request) {
		app.ServeWs(hub, w, r)
	}, authMiddleware)
	r.Handle("GET", "/api/conversations", app.GetConversationsHandler, authMiddleware)
	r.Handle("GET", "/api/messages", app.GetMessagesHandler, authMiddleware)
	r.Handle("GET", "/api/messages/unread-count", app.GetUnreadMessageCountHandler, authMiddleware)
	r.Handle("POST", "/api/messages/read-all", app.MarkAllMessagesAsReadHandler, authMiddleware)
	r.Handle("POST", "/api/profile/privacy", app.UpdateProfilePrivacyHandler, authMiddleware)
	r.Handle("GET", "/api/session/me", app.GetSessionUserHandler, authMiddleware)
	r.Handle("GET", "/api/feed", app.GetFeedHandler, authMiddleware)

	log.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
