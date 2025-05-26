package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	custErr "github.com/PIRSON21/parking/internal/lib/errors"
	"github.com/go-chi/render"
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
func AuthMiddleware(log *slog.Logger, storage AuthGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// читаем session_id из cookie
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Debug("auth middleware", slog.String("session_id", cookie.Value))

			// проверяем сессию в БД
			userID, err := storage.GetUserID(cookie.Value)
			if err != nil {
				log.Error("error while getting userID from storage", slog.String("session_id", cookie.Value), slog.String("err", err.Error()))
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
			log.Debug("userID added to context", slog.Int("userID", userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(UserIDKey)
		if userID, ok := userIDVal.(int); ok {
			if userID != 0 {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, resp.UnknownError("Access denied"))
				return
			}

			next.ServeHTTP(w, r)

			return
		}

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, resp.UnknownError("Unauthorized"))
	})
}

func ManagerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(UserIDKey)
		if userID, ok := userIDVal.(int); ok {
			if userID == 0 {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, resp.UnknownError("Access denied"))
				return
			}

			next.ServeHTTP(w, r)

			return
		}

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, resp.UnknownError("Unauthorized"))
	})
}
