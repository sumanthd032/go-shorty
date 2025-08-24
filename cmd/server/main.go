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
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/handlers"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services"
)
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Database Connection ---
	conn, err := pgx.Connect(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// --- Redis Connection ---
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	// Ping the Redis server to check the connection.
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	// --- Dependency Injection ---
	queries := db.New(conn)
	// Pass the Redis client to the service layer.
	linkService := services.NewLinkService(queries, rdb)
	linkHandler := handlers.NewLinkHandler(linkService)

	// --- Server Setup ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Post("/links", linkHandler.CreateLink)
	})

	// Add the new redirect route at the root level.
	r.Get("/{alias}", linkHandler.Redirect)

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
