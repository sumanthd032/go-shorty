package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services"
)

type UserHandler struct {
	service      *services.UserService
	sessionStore sessions.Store
	queries      *db.Queries
}

func NewUserHandler(s *services.UserService, store sessions.Store, queries *db.Queries) *UserHandler {
	return &UserHandler{service: s, sessionStore: store, queries: queries}
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		// --- FIX: Check for a specific database error ---
		var pgErr *pgconn.PgError
		// Error code '23505' is for unique_violation in PostgreSQL.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, `{"error":"Email address already exists"}`, http.StatusConflict) // 409 Conflict
			return
		}
		// --- END FIX ---

		// For all other errors, return a generic 500.
		http.Error(w, `{"error":"Could not create user"}`, http.StatusInternalServerError)
		return
	}

	userResp := UserResponse{ID: user.ID, Email: user.Email}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	session, _ := h.sessionStore.Get(r, "auth-session")
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		http.Error(w, `{"error":"Could not save session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged in successfully"})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.sessionStore.Get(r, "auth-session")
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, `{"error":"Could not log out"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, `{"error":"User not found in context"}`, http.StatusInternalServerError)
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	userResp := UserResponse{ID: user.ID, Email: user.Email}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
}