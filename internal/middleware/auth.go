package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Auth is a UI-friendly authentication middleware.
func Auth(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "auth-session")

			// Check if the user_id exists in the session.
			userID, ok := session.Values["user_id"].(int64)

			// If the user is not authenticated (id not found or is 0),
			// redirect them to the login page.
			if !ok || userID == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// If they are authenticated, add their ID to the request context
			// and proceed to the next handler (e.g., the dashboard).
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}