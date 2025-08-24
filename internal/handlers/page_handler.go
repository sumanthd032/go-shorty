package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
)

type PageHandler struct {
	templates map[string]*template.Template
	queries   *db.Queries
}

func NewPageHandler(queries *db.Queries) *PageHandler {
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pgUserID := pgtype.Int8{
		Int64: userID,
		Valid: true,
	}

	links, err := h.queries.GetLinksByUserID(r.Context(), pgUserID)
	if err != nil {
		http.Error(w, "Could not fetch links", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Links": links,
	}

	err = h.templates["dashboard.html"].ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("ERROR: Failed to execute dashboard template: %v", err)
		http.Error(w, "Failed to render page", 500)
	}
}