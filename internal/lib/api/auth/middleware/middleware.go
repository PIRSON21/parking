package middleware

import (
	"context"
	"errors"
	"fmt"
	custErr "github.com/PIRSON21/parking/internal/lib/errors"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=AuthGetter
type AuthGetter interface {
	GetUserID(sessionID string) (int, error)
}

type contextKey string

// UserIDKey - ключ для получения userID.
// Используется отдельный тип contextKey, чтобы значение не перекрывалось другими middleware.
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
			if err != nil {
				if errors.Is(err, custErr.ErrUnauthorized) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				} else if errors.Is(err, custErr.ErrSessionExpired) {
					http.Error(w, "Session Expired", http.StatusForbidden)
					return
				}

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(UserIDKey)
		if userID, ok := userIDVal.(int); ok {
			fmt.Println(userID)
			if userID != 0 {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)

			return
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func ManagerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(UserIDKey)
		if userID, ok := userIDVal.(int); ok {
			if userID == 0 {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)

			return
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
