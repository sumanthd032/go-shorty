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
			session, err := store.Get(r, "auth-session")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			userID, ok := session.Values["user_id"].(int64)
			if !ok || userID == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}