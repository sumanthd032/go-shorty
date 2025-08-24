package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware" // Aliased to avoid name conflicts
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	// Make sure your module path is correct
	"github.com/sumanthd032/go-shorty/internal/config"
	"github.com/sumanthd032/go-shorty/internal/handlers"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Connect to Database
	conn, err := pgx.Connect(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// 3. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	// 4. Set up Session Store
	sessionStore := sessions.NewCookieStore([]byte(cfg.Auth.SessionKey))

	// 5. Dependency Injection (Wiring everything together)
	queries := db.New(conn)
	linkService := services.NewLinkService(queries, rdb)
	userService := services.NewUserService(queries)
	
	// Handlers
	linkHandler := handlers.NewLinkHandler(linkService)
	userHandler := handlers.NewUserHandler(userService, sessionStore)
	pageHandler := handlers.NewPageHandler(queries)

	// 6. Set up Middleware
	authMiddleware := middleware.Auth(sessionStore)

	// 7. Set up Router
	r := chi.NewRouter()

	// Use standard middleware from Chi
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// --- Public Routes ---
	r.Get("/{alias}", linkHandler.Redirect) // Short link redirects

	// API routes for login/registration
	r.Post("/api/users/register", userHandler.Register)
	r.Post("/api/users/login", userHandler.Login)
	
	// Page route for the login page
	r.Get("/login", pageHandler.ShowLoginPage)

	// --- Protected Routes (require login) ---
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		// Page routes
		r.Get("/", pageHandler.ShowDashboard)
		r.Post("/ui/links", linkHandler.CreateLinkFromUI)

		// API routes
		r.Post("/api/links", linkHandler.CreateLink)
	})

	// 8. Start the Server
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}