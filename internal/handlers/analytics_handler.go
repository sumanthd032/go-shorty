package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
)

type AnalyticsHandler struct {
	queries *db.Queries
}

func NewAnalyticsHandler(queries *db.Queries) *AnalyticsHandler {
	return &AnalyticsHandler{queries: queries}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, `{"error":"User not authenticated"}`, http.StatusInternalServerError)
		return
	}

	analytics, err := h.queries.GetLinkAnalytics(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		http.Error(w, `{"error":"Could not fetch analytics"}`, http.StatusInternalServerError)
		return
	}

	if analytics == nil {
		analytics = []db.GetLinkAnalyticsRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analytics)
}