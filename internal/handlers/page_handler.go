package handlers

import (
	"html/template"
	"log"
	"net/http"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type PageHandler struct {
	templates map[string]*template.Template
	queries   *db.Queries
}

func NewPageHandler(queries *db.Queries) *PageHandler {
	// Parse all templates on startup
	tmpls, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("could not parse templates: %v", err)
	}

	templatesMap := make(map[string]*template.Template)
	for _, t := range tmpls.Templates() {
		templatesMap[t.Name()] = t
	}

	return &PageHandler{
		templates: templatesMap,
		queries:   queries,
	}
}

func (h *PageHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	h.templates["login.html"].ExecuteTemplate(w, "layout.html", nil)
}

func (h *PageHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		// This should theoretically not happen if middleware is correct
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// We need a new DB query to get links by user ID
	links, err := h.queries.GetLinksByUserID(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		http.Error(w, "Could not fetch links", http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Links": links,
	}

	h.templates["dashboard.html"].ExecuteTemplate(w, "layout.html", data)
}