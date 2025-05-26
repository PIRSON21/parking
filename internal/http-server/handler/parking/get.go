package parking

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/PIRSON21/parking/internal/config"
	authMiddleware "github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	custErr "github.com/PIRSON21/parking/internal/lib/errors"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/xerrors"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=ParkingGetter
type ParkingGetter interface {
	GetAdminParkings(string) ([]*models.Parking, error)
	GetManagerParkings(int, string) ([]*models.Parking, error)
	GetParkingByID(int, int) (*models.Parking, error)
}

// AllParkingsHandler обрабатывает список парковок.
func AllParkingsHandler(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.AllParkingsHandler"

		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var userID int
		var ok bool
		userIDVal := r.Context().Value(authMiddleware.UserIDKey)
		if userID, ok = userIDVal.(int); !ok {
			log.Error("error while getting userID", slog.Any("userID", userIDVal))

			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while getting userID", op))

			return
		}
		log.Debug("userID from context", slog.Int("userID", userID), slog.String("op", op))

		if userID > 0 {
			handleManagerParkings(log, parkingGetter, cfg, w, r, userID)
		} else {
			handleAdminParkings(log, parkingGetter, cfg, w, r)
		}
	}
}

// handleManagerParkings выдает доступные менеджеру парковку.
func handleManagerParkings(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config, w http.ResponseWriter, r *http.Request, userID int) {
	const op = "http-server.handler.parking.handleManagerParkings"

	log = log.With(slog.String("op", op))

	query := r.FormValue("search")
	log.Debug("search query", slog.String("query", query))

	parkings, err := parkingGetter.GetManagerParkings(userID, query)
	if err != nil {
		log.Error("error while getting parkings from DB", slog.String("err", err.Error()))
		resp.ErrorHandler(w, r, cfg, err)
		return
	}

	if len(parkings) == 0 {
		log.Debug("no parkings found for user", slog.Int("userID", userID))
		render.JSON(w, r, []string{})
		return
	}

	log.Debug("found parkings for user", slog.Int("userID", userID), slog.Int("count", len(parkings)))
	if err = render.RenderList(w, r, resp.NewParkingListRender(parkings)); err != nil {
		log.Error("error while marshaling parkings to JSON", slog.String("err", err.Error()))
		resp.ErrorHandler(w, r, cfg, err)
	}
}

// handlerAdminParkings выдает парковки админу.
func handleAdminParkings(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	const op = "http-server.handler.parking.handleAdminParkings"

	log = log.With(slog.String("op", op))

	query := r.FormValue("search")
	log.Debug("search query", slog.String("query", query))

	parkings, err := parkingGetter.GetAdminParkings(query)
	if err != nil {
		log.Error("error while getting from DB",
			slog.String("err", err.Error()))
		resp.ErrorHandler(w, r, cfg, err)
		return
	}

	log.Debug("found parkings for admin", slog.Int("count", len(parkings)))
	if len(parkings) == 0 {
		render.JSON(w, r, []string{})
		return
	}

	if err = render.RenderList(w, r, resp.NewParkingListRender(parkings)); err != nil {
		resp.ErrorHandler(w, r, cfg, err)
		return
	}
}

// GetParkingHandler обрабатывает запрос подробной информации о парковке по его ID.
//
//goland:noinspection ALL
func GetParkingHandler(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.Parking.GetParkingHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		log.Debug("getting parking by ID from request", slog.String("url", r.URL.String()))
		parkingID, err := getParkingID(r)
		if err != nil {
			log.Error("error while getting ID from url", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while getting ID from url: %w", op, err))
			return
		}
		log.Debug("parkingID from url", slog.Int("parkingID", parkingID))

		userID := getUserID(r)
		if userID == -1 {
			log.Error("error while getting userID from context", slog.String("err", "userID not found in context"))
			http.NotFound(w, r)
		}

		log.Debug("userID from context", slog.Int("userID", userID))
		parking, err := parkingGetter.GetParkingByID(parkingID, userID)
		if err != nil {
			if errors.Is(err, custErr.ErrParkingNotFound) {
				log.Debug("parking not found", slog.Int("parkingID", parkingID))
				http.NotFound(w, r)
				return
			} else if errors.Is(err, custErr.ErrParkingAccessDenied) {
				log.Debug("access to parking denied", slog.Int("parkingID", parkingID), slog.Int("userID", userID))
				http.NotFound(w, r)
				return
			}

			log.Error("error while getting Parking from DB", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}

		log.Debug("parking found", slog.Int("parkingID", parking.ID), slog.String("name", parking.Name))
		render.JSON(w, r, parking)
	}
}

// getUserID получает userID о пользователе, полученные из middleware.
func getUserID(r *http.Request) int {
	userIDStr := r.Context().Value(authMiddleware.UserIDKey)
	if userID, ok := userIDStr.(int); ok {
		return userID
	}
	return -1
}

// getParkingID получает ID о парковке из url и проверяет его.
func getParkingID(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return 0, xerrors.Errorf("не указан id парковки")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, xerrors.Errorf("id парковки %v не может быть преобразовано в число", idStr)
	}
	return id, err
}
