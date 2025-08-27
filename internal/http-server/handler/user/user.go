package user

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/PIRSON21/parking/internal/config"
	customMiddleware "github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	"github.com/PIRSON21/parking/internal/lib/api/request"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	customErr "github.com/PIRSON21/parking/internal/lib/errors"
	customValidator "github.com/PIRSON21/parking/internal/lib/validator"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

var InvalidManagerIndex = errors.New("invalid managerID index")

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=UserGetter
type UserGetter interface {
	AuthenticateManager(manager *models.User) (int, error)
	SetSessionID(userID int, sessionID string) error
	GetManagers() ([]*models.User, error)
	GetManagerByID(id int) (*models.User, error)
}

// LoginHandler обрабатывает авторизацию пользователя.
//
//goland:noinspection ALL
func LoginHandler(log *slog.Logger, db UserGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.user.LoginHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("requestID", middleware.GetReqID(r.Context())),
		)

		userReq := new(request.UserLogin)
		err := render.DecodeJSON(r.Body, userReq)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(fmt.Sprintf("error while decoding JSON: %s", err.Error())))
			return
		}

		valid := customValidator.CreateNewValidator()

		if err = valid.Struct(userReq); err != nil {
			validateErr := err.(validator.ValidationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// проверка на администратора. По условиям задачи, администратор должен иметь один единственный аккаунт,
		// которые встроен в коде программы. Я ЗНАЮ, ЧТО ТАК НЕ НАДО. так просили
		if userReq.Email == "admin" && userReq.Password == "admin" {
			log.Debug("admin login attempt")
			err = returnSessionID(w, 0, db)
			if err != nil {
				log.Error("err while returning session ID", slog.String("err", err.Error()))
				resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while returning session: %w", op, err))
				return
			}
			return
		}

		user := &models.User{
			Email:    userReq.Email,
			Password: userReq.Password,
		}

		// проверка введенных данных менеджера
		user.ID, err = db.AuthenticateManager(user)
		if err != nil {
			// в случае, если логин и пароль не найдены или неправильны
			if errors.Is(err, customErr.ErrUnauthorized) {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.UnknownError("неправильный логин или пароль"))

				return
			}
			log.Error("error while getting manager info", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while getting manager info: %w", op, err))

			return
		}
		log.Debug("manager successfully authenticated", slog.Int("userID", user.ID))

		// создание и возврат sessionID в куках
		err = returnSessionID(w, user.ID, db)
		if err != nil {
			log.Error("err while returning session ID", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while returning session: %w", op, err))
			return
		}
		log.Debug("session ID returned successfully", slog.Int("userID", user.ID))
	}
}

// returnSessionID возвращает в куках sessionID в случае удачной авторизации.
func returnSessionID(w http.ResponseWriter, userID int, db UserGetter) error {
	const op = "http-server.handler.user.returnSessionID"
	sessionID := generateSessionID()

	err := db.SetSessionID(userID, sessionID)
	if err != nil {
		return xerrors.Errorf("%s: error while setting session to DB: %w", op, err)
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
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=UserSetter
type UserSetter interface {
	CreateNewManager(*request.UserCreate) error
	UpdateManager(*UserPatch) error
	DeleteManager(int) error
	GetManagerByID(id int) (*models.User, error)
}

// CreateManagerHandler обрабатывает запрос на создание менеджера
func CreateManagerHandler(log *slog.Logger, db UserSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.user.CreateManagerHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		newManager := new(request.UserCreate)
		if err := render.DecodeJSON(r.Body, newManager); err != nil {
			log.Error("error while decoding JSON", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("new manager from request", slog.Any("newManager", newManager))

		valid := customValidator.CreateNewValidator()
		if err := valid.Struct(newManager); err != nil {
			log.Error("validation error", slog.String("err", err.Error()))
			err := err.(validator.ValidationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(err))
			return
		}

		log.Debug("creating new manager in DB", slog.Any("newManager", newManager))
		err := db.CreateNewManager(newManager)
		if err != nil {
			if errors.Is(err, customErr.ErrManagerAlreadyExists) {
				log.Error("manager already exists", slog.String("err", err.Error()))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, resp.UnknownError(err.Error()))
				return
			}
			log.Error("error while creating new manager", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("new manager created successfully", slog.Any("newManager", newManager))

		w.WriteHeader(http.StatusCreated)
	}
}

// GetManagersHandler выдает всех менеджеров.
func GetManagersHandler(log *slog.Logger, db UserGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.user.GetManagersHandler"

		log = log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
			slog.String("op", op),
		)

		log.Debug("getting all managers from DB")
		managers, err := db.GetManagers()
		if err != nil {
			log.Error("error while getting managers", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}

		if managers == nil {
			log.Debug("no managers found")
			render.JSON(w, r, []string{})
			return
		}

		log.Debug("found managers", slog.Int("count", len(managers)))
		if err := render.RenderList(w, r, resp.NewManagerListRender(managers)); err != nil {
			log.Error("error while rendering managers", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
	}
}

// GetManagerByIDHandler выдает полную информацию о менеджере по его ID.
func GetManagerByIDHandler(log *slog.Logger, db UserGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.user.GetManagerByIDHandler"

		log = log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		managerID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(InvalidManagerIndex.Error()))
			return
		}
		log.Debug("managerID from URL", slog.Int("managerID", managerID), slog.String("op", op))

		getManagerAndSend(w, r, db, managerID, cfg, log)
	}
}

type UserGetterByID interface {
	GetManagerByID(int) (*models.User, error)
}

func getManagerAndSend(w http.ResponseWriter, r *http.Request, db UserGetterByID, managerID int, cfg *config.Config, log *slog.Logger) {
	const op = "handler.user.getManagerAndSend"

	log = log.With(slog.String("op", op))

	log.Debug("getting manager by ID", slog.Int("managerID", managerID))
	manager, err := db.GetManagerByID(managerID)
	if err != nil {
		if errors.Is(err, customErr.ErrManagerNotFound) {
			log.Error("manager not found", slog.Int("managerID", managerID))
			http.NotFound(w, r)
			return
		}

		log.Error("error while getting manager by ID", slog.String("err", err.Error()))
		resp.ErrorHandler(w, r, cfg, err)
		return
	}

	log.Debug("manager found", slog.Int("managerID", manager.ID), slog.String("login", manager.Login))

	render.JSON(w, r, resp.NewManagerResponse(manager))
}

type UserPatch struct {
	ID       int
	Login    *string `json:"login,omitempty" validate:"omitempty,min=4,max=8"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=4,max=10"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email,min=8,max=15"`
}

// UpdateManagerHandler обновляет данные о менеджере.
func UpdateManagerHandler(log *slog.Logger, db UserSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.user.UpdateManagerHandler"

		log = log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		managerID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("error while getting ID", slog.String("err", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(InvalidManagerIndex.Error()))
			return
		}
		log.Debug("managerID from URL", slog.Int("managerID", managerID), slog.String("op", op))

		var managerUpdate UserPatch
		err = render.DecodeJSON(r.Body, &managerUpdate)
		managerUpdate.ID = managerID
		if err != nil {
			log.Error("error while decoding JSON", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("manager update request", slog.Any("managerUpdate", managerUpdate), slog.String("op", op))

		if managerUpdate.Email == nil && managerUpdate.Login == nil && managerUpdate.Password == nil {
			log.Error("no data provided for update", slog.String("op", op))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError("no data provided"))
			return
		}

		valid := customValidator.CreateNewValidator()
		if err := valid.Struct(managerUpdate); err != nil {
			var validErr validator.ValidationErrors
			if ok := errors.As(err, &validErr); ok {
				log.Error("validation error", slog.String("err", err.Error()), slog.String("op", op))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ValidationError(validErr))
				return
			}
			log.Error("error while validating struct", slog.String("err", err.Error()), slog.String("op", op))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}

		log.Debug("updating manager in DB", slog.Any("managerUpdate", managerUpdate), slog.String("op", op))
		err = db.UpdateManager(&managerUpdate)
		if err != nil {
			log.Error("error while updating manager", slog.String("err", err.Error()), slog.String("op", op))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("manager updated successfully", slog.Int("managerID", managerID), slog.String("op", op))

		getManagerAndSend(w, r, db, managerID, cfg, log)
	}
}

// DeleteManagerHandler удаляет менеджера.
func DeleteManagerHandler(log *slog.Logger, db UserSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.user.DeleteManagerHandler"

		log = log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
			slog.String("op", op),
		)

		managerID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("error while parsing ID", slog.String("err", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(InvalidManagerIndex.Error()))
			return
		}
		log.Debug("managerID from URL", slog.Int("managerID", managerID))

		err = db.DeleteManager(managerID)
		if err != nil {
			log.Error("error while deleting manager", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("manager deleted successfully", slog.Int("managerID", managerID))

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetRoleHandler получает роль пользователя, если он авторизован.
func GetRoleHandler(log *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.user.GetRoleHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		tmp := r.Context().Value(customMiddleware.UserIDKey)
		userID, ok := tmp.(int)
		if !ok {
			log.Error("error while converting userID to int", slog.Any("usedID", tmp))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error with userID: %q", op, tmp))
			return
		}
		log.Debug("userID from context", slog.Int("userID", userID))

		if userID == 0 {
			render.JSON(w, r, map[string]interface{}{
				"userID": 0,
				"role":   "admin",
			})
			return
		} else if userID > 0 {
			render.JSON(w, r, map[string]interface{}{
				"userID": userID,
				"role":   "manager",
			})
			return
		} else {
			log.Error("invalid userID", slog.Int("userID", userID))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: invalid userID: %d", op, userID))
			return
		}
	}
}
