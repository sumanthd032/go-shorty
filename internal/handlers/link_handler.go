// File: internal/handlers/link_handler.go

package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sumanthd032/go-shorty/internal/services"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"html/template" 
	
)

// LinkHandler handles HTTP requests for links.
type LinkHandler struct {
	service *services.LinkService
	templates map[string]*template.Template 
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(s *services.LinkService) *LinkHandler {
	tmpls, _ := template.ParseGlob("templates/*.html")
	templatesMap := make(map[string]*template.Template)
	for _, t := range tmpls.Templates() {
		templatesMap[t.Name()] = t
	}
	return &LinkHandler{service: s, templates: templatesMap}
}


// CreateLinkRequest defines the expected JSON body for creating a link.
type CreateLinkRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"` // omitempty means the field is optional
}

func (h *LinkHandler) CreateLinkFromUI(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	// We parse form data instead of JSON
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	originalURL := r.FormValue("url")
	customAlias := r.FormValue("alias")
	
	params := services.CreateLinkParams{
		OriginalURL: originalURL,
		CustomAlias: customAlias,
		UserID:      userID,
	}

	link, err := h.service.Create(r.Context(), params)
	if err != nil {
		http.Error(w, "Could not create link", http.StatusInternalServerError)
		return
	}
	
	// Instead of sending JSON, we execute the link_row template with the new link data
	h.templates["link_row.html"].Execute(w, link)
}

// CreateLink is the handler for the POST /api/links endpoint.
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusInternalServerError)
		return
	}

	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	params := services.CreateLinkParams{
		OriginalURL: req.URL,
		CustomAlias: req.Alias,
		UserID:      userID, // Pass the user ID
	}

	link, err := h.service.Create(r.Context(), params)
	if err != nil {
		if errors.Is(err, services.ErrAliasExists) {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		log.Printf("Internal server error: %v", err)
		http.Error(w, "Could not create link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(link)
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		http.Error(w, "Alias is missing", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURL(r.Context(), alias)
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			http.NotFound(w, r) // Renders a standard 404 page
			return
		}
		log.Printf("Internal server error on redirect: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Perform the redirect
	http.Redirect(w, r, originalURL, http.StatusFound) // 302 Found
}