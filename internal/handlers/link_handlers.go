// File: internal/handlers/link_handler.go

package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sumanthd032/go-shorty/internal/services"
)

// LinkHandler handles HTTP requests for links.
type LinkHandler struct {
	service *services.LinkService
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(s *services.LinkService) *LinkHandler {
	return &LinkHandler{service: s}
}

// CreateLinkRequest defines the expected JSON body for creating a link.
type CreateLinkRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"` // omitempty means the field is optional
}

// CreateLink is the handler for the POST /api/links endpoint.
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
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