
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sumanthd032/go-shorty/internal/config"
)

func main() {
	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Go-Shorty!"))
	})

	// Use the port from our loaded configuration
	// fmt.Sprintf formats the string, replacing %d with the value of cfg.Server.Port
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}