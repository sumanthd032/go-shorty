// File: cmd/server/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/config"
	"github.com/sumanthd032/go-shorty/internal/handlers"
	authMiddleware "github.com/sumanthd032/go-shorty/internal/middleware" 
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
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	// --- Session Store ---
	sessionStore := sessions.NewCookieStore([]byte(cfg.Auth.SessionKey))

	// --- Dependency Injection ---
	queries := db.New(conn)
	linkService := services.NewLinkService(queries, rdb)
	userService := services.NewUserService(queries)
	linkHandler := handlers.NewLinkHandler(linkService)
	userHandler := handlers.NewUserHandler(userService, sessionStore)

	// --- Middleware ---
	// Calling our custom middleware from the renamed package import
	requireAuth := authMiddleware.Auth(sessionStore)

	// --- Server Setup ---
	r := chi.NewRouter()

	// FIX: These now correctly refer to the chi/v5/middleware package
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/{alias}", linkHandler.Redirect)
	r.Post("/api/users/register", userHandler.Register)
	r.Post("/api/users/login", userHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(requireAuth) // Use our auth middleware
		r.Post("/api/links", linkHandler.CreateLink)
	})

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}