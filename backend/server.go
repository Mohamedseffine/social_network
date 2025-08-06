package main

import (
	"encoding/json"
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/pkg/handlers"

	"github.com/gorilla/mux"
)

func main() {
	db, err := sqlite.InitDB("social-network.db", "pkg/db/migrations/sqlite")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	env := &handlers.Env{DB: db}

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// Public routes
	api.HandleFunc("/register", env.RegisterHandler).Methods("POST")
	api.HandleFunc("/login", env.LoginHandler).Methods("POST")
	api.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	}).Methods("GET")
	api.HandleFunc("/users/{id}/followers", env.ListFollowersHandler).Methods("GET")
	api.HandleFunc("/users/{id}/following", env.ListFollowingHandler).Methods("GET")

	// Authenticated routes
	auth := api.PathPrefix("").Subrouter()
	auth.Use(env.AuthMiddleware)
	auth.HandleFunc("/logout", env.LogoutHandler).Methods("POST")
	auth.HandleFunc("/users/{id}/follow", env.FollowUserHandler).Methods("POST")
	auth.HandleFunc("/follow-requests/{id}", env.HandleFollowRequestHandler).Methods("POST")
	auth.HandleFunc("/users/{id}/profile", env.GetProfileHandler).Methods("GET")
	auth.HandleFunc("/profile", env.UpdateProfileHandler).Methods("PUT")

	log.Println("Server starting on port 8081...")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
