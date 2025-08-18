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

	apiRouter := router.NewRouter()

	// Auth routes
	apiRouter.HandleFunc("/register", app.RegisterHandler).Methods("POST")
	apiRouter.HandleFunc("/login", app.LoginHandler).Methods("POST")
	apiRouter.HandleFunc("/upload", app.UploadImageHandler).Methods("POST")

	// Authenticated routes
	auth := func(next http.Handler) http.Handler {
		return app.Authenticate(next)
	}

	apiRouter.Handle("/logout", auth(http.HandlerFunc(app.LogoutHandler))).Methods("POST")
	apiRouter.Handle("/users/{id}/follow", auth(http.HandlerFunc(app.FollowUserHandler))).Methods("POST")
	apiRouter.Handle("/users/{id}/unfollow", auth(http.HandlerFunc(app.UnfollowUserHandler))).Methods("POST")
	apiRouter.Handle("/requests/{id}/accept", auth(http.HandlerFunc(app.AcceptFollowRequestHandler))).Methods("POST")
	apiRouter.Handle("/requests/{id}/decline", auth(http.HandlerFunc(app.DeclineFollowRequestHandler))).Methods("POST")
	apiRouter.Handle("/users", auth(http.HandlerFunc(app.GetAllUsersHandler))).Methods("GET")
	apiRouter.Handle("/users/{id}", auth(http.HandlerFunc(app.GetUserHandler))).Methods("GET")
	apiRouter.Handle("/users/online", auth(http.HandlerFunc(app.GetOnlineUsersHandler))).Methods("GET")
	apiRouter.Handle("/users/{id}/followers", auth(http.HandlerFunc(app.GetFollowersHandler))).Methods("GET")
	apiRouter.Handle("/users/{id}/following", auth(http.HandlerFunc(app.GetFollowingHandler))).Methods("GET")
	apiRouter.Handle("/users/{id}/posts", auth(http.HandlerFunc(app.GetUserPostsHandler))).Methods("GET")
	apiRouter.Handle("/posts", auth(http.HandlerFunc(app.CreatePostHandler))).Methods("POST")
	// Comments routes
	apiRouter.Handle("/posts/{id}/comments", auth(http.HandlerFunc(app.CreateCommentHandler))).Methods("POST")
	apiRouter.Handle("/posts/{id}/comments", auth(http.HandlerFunc(app.GetCommentsHandler))).Methods("GET")

	// Group routes
	apiRouter.Handle("/groups", auth(http.HandlerFunc(app.CreateGroupHandler))).Methods("POST")
	apiRouter.Handle("/groups", auth(http.HandlerFunc(app.GetGroupsHandler))).Methods("GET")
	apiRouter.Handle("/groups/{id}", auth(http.HandlerFunc(app.GetGroupHandler))).Methods("GET")
	apiRouter.Handle("/groups/{id}/membership", auth(http.HandlerFunc(app.GetGroupMembershipStatusHandler))).Methods("GET")
	apiRouter.Handle("/groups/{id}/join", auth(http.HandlerFunc(app.JoinGroupHandler))).Methods("POST")
	apiRouter.Handle("/groups/requests/{id}/accept", auth(http.HandlerFunc(app.AcceptGroupRequestHandler))).Methods("POST")
	apiRouter.Handle("/groups/requests/{id}/decline", auth(http.HandlerFunc(app.DeclineGroupRequestHandler))).Methods("POST")
	apiRouter.Handle("/groups/{id}/invite", auth(http.HandlerFunc(app.InviteToGroupHandler))).Methods("POST")
	apiRouter.Handle("/groups/invites/{id}/accept", auth(http.HandlerFunc(app.AcceptGroupInviteHandler))).Methods("POST")
	apiRouter.Handle("/groups/invites/{id}/decline", auth(http.HandlerFunc(app.DeclineGroupInviteHandler))).Methods("POST")
	apiRouter.Handle("/groups/{id}/posts", auth(http.HandlerFunc(app.CreateGroupPostHandler))).Methods("POST")
	apiRouter.Handle("/groups/{id}/posts", auth(http.HandlerFunc(app.GetGroupPostsHandler))).Methods("GET")
	apiRouter.Handle("/group-posts/{id}/comments", auth(http.HandlerFunc(app.CreateGroupPostCommentHandler))).Methods("POST")
	apiRouter.Handle("/group-posts/{id}/comments", auth(http.HandlerFunc(app.GetGroupPostCommentsHandler))).Methods("GET")
	apiRouter.Handle("/search_groups", auth(http.HandlerFunc(app.SearchGroupsHandler))).Methods("GET")

	// Event routes
	apiRouter.Handle("/groups/{id}/events", auth(http.HandlerFunc(app.CreateEventHandler))).Methods("POST")
	apiRouter.Handle("/groups/{id}/events", auth(http.HandlerFunc(app.GetGroupEventsHandler))).Methods("GET")
	apiRouter.Handle("/events/{id}/respond", auth(http.HandlerFunc(app.RespondToEventHandler))).Methods("POST")
	apiRouter.Handle("/events/{id}/attendees", auth(http.HandlerFunc(app.GetEventAttendeesHandler))).Methods("GET")

	// Notification routes
	apiRouter.Handle("/notifications", auth(http.HandlerFunc(app.GetNotificationsHandler))).Methods("GET")
	apiRouter.Handle("/notifications/unread-count", auth(http.HandlerFunc(app.GetUnreadNotificationCountHandler))).Methods("GET")
	apiRouter.Handle("/notifications/read-all", auth(http.HandlerFunc(app.MarkAllNotificationsAsReadHandler))).Methods("POST")
	apiRouter.Handle("/notifications/{id}/read", auth(http.HandlerFunc(app.MarkNotificationAsReadHandler))).Methods("POST")
	apiRouter.Handle("/notifications/{id}", auth(http.HandlerFunc(app.DeleteNotificationHandler))).Methods("DELETE")

	// WebSocket route
	apiRouter.Handle("/ws", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.ServeWs(hub, w, r)
	})))

	// Chat routes
	apiRouter.Handle("/conversations", auth(http.HandlerFunc(app.GetConversationsHandler))).Methods("GET")
	apiRouter.Handle("/messages", auth(http.HandlerFunc(app.GetMessagesHandler))).Methods("GET")
	apiRouter.Handle("/messages/unread-count", auth(http.HandlerFunc(app.GetUnreadMessageCountHandler))).Methods("GET")
	apiRouter.Handle("/messages/read-all", auth(http.HandlerFunc(app.MarkAllMessagesAsReadHandler))).Methods("POST")

	// Profile privacy route
	apiRouter.Handle("/profile/privacy", auth(http.HandlerFunc(app.UpdateProfilePrivacyHandler))).Methods("POST")

	// Session route
	apiRouter.Handle("/session/me", auth(http.HandlerFunc(app.GetSessionUserHandler))).Methods("GET")

	// Feed route
	apiRouter.Handle("/feed", auth(http.HandlerFunc(app.GetFeedHandler))).Methods("GET")

	// Main router
	mainRouter := http.NewServeMux()
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mainRouter.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Apply CORS middleware
	corsHandler := handlers.CORS(mainRouter)

	log.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", corsHandler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
