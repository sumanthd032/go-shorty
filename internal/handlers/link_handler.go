package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services"
)

type LinkHandler struct {
	service *services.LinkService
	queries *db.Queries // Add queries to fetch links
}

func NewLinkHandler(s *services.LinkService, queries *db.Queries) *LinkHandler {
	return &LinkHandler{service: s, queries: queries}
}

type CreateLinkRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, `{"error":"User not authenticated"}`, http.StatusInternalServerError)
		return
	}

	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, `{"error":"URL is required"}`, http.StatusBadRequest)
		return
	}

	params := services.CreateLinkParams{
		OriginalURL: req.URL,
		CustomAlias: req.Alias,
		UserID:      userID,
	}

	link, err := h.service.Create(r.Context(), params)
	if err != nil {
		if errors.Is(err, services.ErrAliasExists) {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
			return
		}
		log.Printf("Internal server error: %v", err)
		http.Error(w, `{"error":"Could not create link"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		http.Error(w, "Alias is missing", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURLAndTrack(r.Context(), alias, r.RemoteAddr, r.UserAgent(), r.Referer())
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("Internal server error on redirect: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// GetUserLinks is the handler for the GET /api/links endpoint.
func (h *LinkHandler) GetUserLinks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, `{"error":"User not authenticated"}`, http.StatusInternalServerError)
		return
	}

	links, err := h.queries.GetLinksByUserID(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		http.Error(w, `{"error":"Could not fetch links"}`, http.StatusInternalServerError)
		return
	}

	// Handle case where user has no links
	if links == nil {
		links = []db.Link{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(links)
}