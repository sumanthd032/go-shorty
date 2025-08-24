// File: cmd/server/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/sumanthd032/go-shorty/internal/config"
	// Import the generated db package
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Connect to the database
	// We use context.Background() as a default, empty context.
	conn, err := pgx.Connect(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	// Defer closing the connection until the main function exits.
	defer conn.Close(context.Background())

	// 3. Create a new Querier instance from our generated code
	queries := db.New(conn)

	// --- Server Setup ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Go-Shorty! DB connection successful."))
	})

    // Here we could pass `queries` to our handlers and services

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}