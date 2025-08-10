package main

import (
	"log"
	"net/http"
	"social-network/backend/pkg/db/sqlite"
	"social-network/backend/pkg/handlers"
	"social-network/backend/pkg/websockets"

	"github.com/gorilla/mux"
)

func main() {
	db := sqlite.InitDB("social-network.db")
	defer db.Close()

	hub := websockets.NewHub(db)
	go hub.Run()
	app := &handlers.App{DB: db, Hub: hub}

	r := mux.NewRouter()

	// Use CORS middleware for all routes
	r.Use(handlers.CORS)

	// Subrouter for API
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Auth routes
	apiRouter.HandleFunc("/register", app.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", app.LoginHandler).Methods("POST", "OPTIONS")

	// Authenticated routes
	authRouter := apiRouter.PathPrefix("/").Subrouter()
	authRouter.Use(app.Authenticate)
	authRouter.HandleFunc("/logout", app.LogoutHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/users/{id}/follow", app.FollowUserHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/users/{id}/unfollow", app.UnfollowUserHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/requests/{id}/accept", app.AcceptFollowRequestHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/requests/{id}/decline", app.DeclineFollowRequestHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/users", app.GetAllUsersHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/users/{id}", app.GetUserHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/users/{id}/followers", app.GetFollowersHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/users/{id}/following", app.GetFollowingHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/users/{id}/posts", app.GetUserPostsHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/posts", app.CreatePostHandler).Methods("POST", "OPTIONS")

	// Group routes
	authRouter.HandleFunc("/groups", app.CreateGroupHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups", app.GetGroupsHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}", app.GetGroupHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/membership", app.GetGroupMembershipStatusHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/join", app.JoinGroupHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/requests/{id}/accept", app.AcceptGroupRequestHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/requests/{id}/decline", app.DeclineGroupRequestHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/invite", app.InviteToGroupHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/invites/{id}/accept", app.AcceptGroupInviteHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/invites/{id}/decline", app.DeclineGroupInviteHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/posts", app.CreateGroupPostHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/posts", app.GetGroupPostsHandler).Methods("GET", "OPTIONS")

	// Event routes
	authRouter.HandleFunc("/groups/{id}/events", app.CreateEventHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/groups/{id}/events", app.GetGroupEventsHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/events/{id}/respond", app.RespondToEventHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/events/{id}/attendees", app.GetEventAttendeesHandler).Methods("GET", "OPTIONS")

	// Notification routes
	authRouter.HandleFunc("/notifications", app.GetNotificationsHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/notifications/{id}/read", app.MarkNotificationAsReadHandler).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/notifications/{id}", app.DeleteNotificationHandler).Methods("DELETE", "OPTIONS")

	// WebSocket route
	authRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		app.ServeWs(hub, w, r)
	})

	// Chat routes
	authRouter.HandleFunc("/conversations", app.GetConversationsHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/messages", app.GetMessagesHandler).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/messages/unread-count", app.GetUnreadMessageCountHandler).Methods("GET", "OPTIONS")

	// Image upload route
	authRouter.HandleFunc("/upload", app.UploadImageHandler).Methods("POST", "OPTIONS")

	// Profile privacy route
	authRouter.HandleFunc("/profile/privacy", app.UpdateProfilePrivacyHandler).Methods("POST", "OPTIONS")

	// Session route
	authRouter.HandleFunc("/session/me", app.GetSessionUserHandler).Methods("GET", "OPTIONS")

	// Feed route
	authRouter.HandleFunc("/feed", app.GetFeedHandler).Methods("GET", "OPTIONS")

	// Serve static files
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	log.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
