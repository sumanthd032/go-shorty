package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sumanthd032/go-shorty/internal/middleware"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
)

// AnalyticsResponse defines the JSON structure for the analytics data.
// The json tags ensure the output matches what the JavaScript expects.
type AnalyticsResponse struct {
	ID          int64  `json:"id"`
	Alias       string `json:"alias"`
	OriginalURL string `json:"original_url"`
	TotalClicks int64  `json:"total_clicks"`
}

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

	dbAnalytics, err := h.queries.GetLinkAnalytics(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		http.Error(w, `{"error":"Could not fetch analytics"}`, http.StatusInternalServerError)
		return
	}

	// Convert the database models to our API response models.
	apiAnalytics := make([]AnalyticsResponse, 0, len(dbAnalytics))
	for _, item := range dbAnalytics {
		apiAnalytics = append(apiAnalytics, AnalyticsResponse{
			ID:          item.ID,
			Alias:       item.Alias,
			OriginalURL: item.OriginalUrl,
			TotalClicks: item.TotalClicks,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiAnalytics)
}