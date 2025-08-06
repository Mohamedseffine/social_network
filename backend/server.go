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

	r.HandleFunc("/api/register", env.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", env.LoginHandler).Methods("POST")
	r.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	}).Methods("GET")

	log.Println("Server starting on port 8081...")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
