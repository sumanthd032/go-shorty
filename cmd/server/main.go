package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Create a new chi router.
	// A router is responsible for directing incoming web requests
	// to the correct handler function.
	r := chi.NewRouter()

	// --- Middleware ---
	// Middleware are functions that run on every request before it
	// reaches its final destination (our handler). They are great for
	// cross-cutting concerns like logging, handling panics, etc.

	// This middleware logs the start and end of each request,
	// along with some useful information like the URL and processing time.
	r.Use(middleware.Logger)
	// This middleware recovers from panics anywhere in the handler chain,
	// so that your server doesn't crash. It will return a 500 Internal Server Error.
	r.Use(middleware.Recoverer)

	// --- Routes ---
	// A route defines a URL pattern and the handler function that
	// should be executed when a request matches that pattern.

	// We are defining a route for the root URL ("/").
	// When a GET request comes in for "/", the anonymous function we've
	// defined will be executed.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// w http.ResponseWriter is what we use to write our response back to the user.
		// r *http.Request contains all the information about the incoming request.
		w.Write([]byte("Welcome to Go-Shorty!"))
	})

	// --- Starting the Server ---
	// We define the address the server will listen on. ":8080" means it will
	// listen on all available network interfaces on port 8080.
	const port = ":8080"
	log.Printf("Server starting on port %s", port)

	// http.ListenAndServe starts the HTTP server.
	// It takes the address and the router as arguments.
	// This function will block indefinitely, or until the server
	// encounters an unrecoverable error.
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}