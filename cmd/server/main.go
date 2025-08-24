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
	// Import our new packages
	"github.com/sumanthd032/go-shorty/internal/handlers"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conn, err := pgx.Connect(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// --- Dependency Injection ---
	// We create our dependencies starting from the innermost layer (repository)
	// and working our way out.
	queries := db.New(conn)
	linkService := services.NewLinkService(queries)
	linkHandler := handlers.NewLinkHandler(linkService)

	// --- Server Setup ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Create a new router group for API endpoints
	r.Route("/api", func(r chi.Router) {
		r.Post("/links", linkHandler.CreateLink)
	})

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}