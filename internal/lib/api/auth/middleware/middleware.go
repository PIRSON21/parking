package middleware

import (
	"context"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=AuthGetter
type AuthGetter interface {
	GetUserID(sessionID string) (int, error)
}

type contextKey string

var UserIDKey contextKey = "userID"

// AuthMiddleware проверяет session_id из cookie клиента на актуальность и достоверность.
func AuthMiddleware(storage AuthGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// читаем session_id из cookie
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// проверяем сессию в БД
			userID, err := storage.GetUserID(cookie.Value)
			if userID == -1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
