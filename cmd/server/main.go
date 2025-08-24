package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/config"
	"github.com/sumanthd032/go-shorty/internal/handlers"
	"github.com/sumanthd032/go-shorty/internal/middleware"
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	sessionStore := sessions.NewCookieStore([]byte(cfg.Auth.SessionKey))
	queries := db.New(conn)

	linkService := services.NewLinkService(queries, rdb)
	userService := services.NewUserService(queries)

	linkHandler := handlers.NewLinkHandler(linkService, queries)
	userHandler := handlers.NewUserHandler(userService, sessionStore, queries)
	analyticsHandler := handlers.NewAnalyticsHandler(queries)

	authMiddleware := middleware.Auth(sessionStore)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- API Routes ---
	r.Route("/api", func(r chi.Router) {
		r.Post("/users/register", userHandler.Register)
		r.Post("/users/login", userHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/users/logout", userHandler.Logout)
			r.Post("/links", linkHandler.CreateLink)
			r.Get("/links", linkHandler.GetUserLinks)
			r.Get("/users/me", userHandler.GetCurrentUser)
			r.Get("/analytics", analyticsHandler.GetAnalytics)
		})
	})

	// --- Public Redirect Route ---
	// FIX: Added a regex to the alias to prevent it from matching files with extensions (like .html).
	// It now only matches aliases containing letters, numbers, underscores, and hyphens.
	r.Get("/{alias:[a-zA-Z0-9_-]+}", linkHandler.Redirect)

	// --- Static File Server for the UI ---
	// This will now correctly handle requests for .html files because the route above no longer intercepts them.
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	r.Handle("/*", http.StripPrefix("/", http.FileServer(filesDir)))

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s, serving UI from ./static/", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}