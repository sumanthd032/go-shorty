package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "auth-session")
			userID, ok := session.Values["user_id"].(int64)

			if !ok || userID == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}