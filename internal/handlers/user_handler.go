package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/sessions"
	"github.com/sumanthd032/go-shorty/internal/services"
)

type UserHandler struct {
	service     *services.UserService
	sessionStore sessions.Store
}

func NewUserHandler(s *services.UserService, store sessions.Store) *UserHandler {
	return &UserHandler{service: s, sessionStore: store}
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user) // Note: Don't send password hash in a real app
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
    // Check content type to decide how to parse
	var email, password string
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		var req UserRequest
		json.NewDecoder(r.Body).Decode(&req)
		email = req.Email
		password = req.Password
	} else {
		r.ParseForm()
		email = r.FormValue("email")
		password = r.FormValue("password")
	}
    
    user, err := h.service.Login(r.Context(), email, password)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	session, _ := h.sessionStore.Get(r, "auth-session")
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Could not save session", http.StatusInternalServerError)
		return
	}
    
    // If the request was from a form, redirect to the dashboard
    if contentType != "application/json" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in successfully"))
}