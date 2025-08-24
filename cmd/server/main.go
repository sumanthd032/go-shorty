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
	pageHandler := handlers.NewPageHandler(queries)

	// --- Middleware ---
	// Calling our custom middleware from the renamed package import
	requireAuth := authMiddleware.Auth(sessionStore)

	// --- Server Setup ---
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // --- Page Routes ---
    r.Get("/login", pageHandler.ShowLoginPage)
    
    // Group for routes that require authentication
    r.Group(func(r chi.Router) {
        r.Use(requireAuth)
        r.Get("/", pageHandler.ShowDashboard)
        // This is the new endpoint our HTMX form will post to
        r.Post("/ui/links", linkHandler.CreateLinkFromUI)
    })
    
    // --- API Routes (unchanged) ---
    r.Post("/api/users/register", userHandler.Register)
    r.Post("/api/users/login", userHandler.Login)
    r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Post("/api/links", linkHandler.CreateLink) // The original API endpoint
	})

    // Public redirect route
    r.Get("/{alias}", linkHandler.Redirect)

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}