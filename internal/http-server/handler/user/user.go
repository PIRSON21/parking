package user

import (
	"errors"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	customErr "github.com/PIRSON21/parking/internal/lib/errors"
	customValidator "github.com/PIRSON21/parking/internal/lib/validator"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
	"net/http"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=UserGetter
type UserGetter interface {
	AuthenticateManager(manager *models.User) (int, error)
	SetSessionID(userID int, sessionID string) error
}

// LoginHandler обрабатывает авторизацию пользователя.
//
//goland:noinspection ALL
func LoginHandler(log *slog.Logger, db UserGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.user.LoginHandler"

		log.With(
			slog.String("op", op),
			slog.String("requestID", middleware.GetReqID(r.Context())),
		)

		var user models.User
		err := render.DecodeJSON(r.Body, &user)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(fmt.Sprintf("error while decoding JSON: %s", err.Error())))

			return
		}

		valid := customValidator.CreateNewValidator()

		if err = valid.Struct(&user); err != nil {
			validateErr := err.(validator.ValidationErrors)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		fmt.Println(user)

		// проверка на администратора. По условиям задачи, администратор должен иметь один единственный аккаунт,
		// которые встроен в коде программы
		if user.Login == "admin" && user.Password == "admin" {
			err = returnSessionID(w, 0, db)
			if err != nil {
				log.Error("err while returning session ID", slog.String("err", err.Error()))

				resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while returning session: %w", op, err))

				return
			}

			return
		}

		// проверка введенных данных менеджера
		user.ID, err = db.AuthenticateManager(&user)
		if err != nil {
			// в случае, если логин и пароль не найдены или неправильны
			if errors.Is(err, customErr.ErrUnauthorized) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.UnknownError("неправильный логин или пароль"))

				return
			}
			log.Error("error while getting manager info", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while getting manager info: %w", op, err))

			return
		}

		// создание и возврат sessionID в куках
		err = returnSessionID(w, user.ID, db)
		if err != nil {
			log.Error("err while returning session ID", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while returning session: %w", op, err))

			return
		}
	}
}

// returnSessionID возвращает в куках sessionID в случае удачной авторизации.
func returnSessionID(w http.ResponseWriter, userID int, db UserGetter) error {
	const op = "http-server.handler.user.returnSessionID"
	sessionID := generateSessionID()

	err := db.SetSessionID(userID, sessionID)
	if err != nil {
		return fmt.Errorf("%s: error while setting session to DB: %w", op, err)
	}

	setSessionCookie(w, sessionID)

	w.WriteHeader(http.StatusAccepted)

	return nil
}

// generateSessionID генерирует случайное uuid для сессии.
func generateSessionID() string {
	return uuid.New().String()
}

// setSessionCookie создает cookie для ответа.
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(48 * time.Hour),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, &cookie)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=UserSetter
type UserSetter interface {
	CreateNewManager(*models.User) error
}

// CreateManagerHandler обрабатывает запрос на создание менеджера
func CreateManagerHandler(log *slog.Logger, db UserSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.user.CreateManagerHandler"

		log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		newManager := new(models.User)
		if err := render.DecodeJSON(r.Body, newManager); err != nil {
			log.Error("error while decoding JSON", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, err)

			return
		}

		valid := customValidator.CreateNewValidator()
		if err := valid.Struct(newManager); err != nil {
			err := err.(validator.ValidationErrors)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(err))

			return
		}

		err := db.CreateNewManager(newManager)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, resp.UnknownError(customErr.ErrManagerAlreadyExists.Error()))

				return
			}

			log.Error("error while creating new manager", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, err)

			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
