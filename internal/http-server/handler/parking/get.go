package parking

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	authMiddleware "github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
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

			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while getting userID", op))

			return
		}

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

	log.With(slog.String("op", op))

	query := r.FormValue("search")

	parkings, err := parkingGetter.GetManagerParkings(userID, query)
	if err != nil {
		log.Error("error while getting parkings from DB", slog.String("err", err.Error()))

		resp.ErrorHandler(w, r, cfg, err)

		return
	}

	if len(parkings) == 0 {
		render.JSON(w, r, []string{})

		return
	}

	if err = render.RenderList(w, r, resp.NewParkingListRender(parkings)); err != nil {
		log.Error("error while marshaling parkings to JSON", slog.String("err", err.Error()))

		resp.ErrorHandler(w, r, cfg, err)
	}
}

// handlerAdminParkings выдает парковки админу.
func handleAdminParkings(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	const op = "http-server.handler.parking.handleAdminParkings"

	log.With(slog.String("op", op))

	query := r.FormValue("search")

	parkings, err := parkingGetter.GetAdminParkings(query)
	if err != nil {
		log.Error("error while getting from DB",
			slog.String("err", err.Error()))

		resp.ErrorHandler(w, r, cfg, err)

		return
	}

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

		log := log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		parkingID, err := getParkingID(r)
		if err != nil {
			log.Error("error while getting ID from url", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while getting ID from url: %w", op, err))
			return
		}

		userID := getUserID(r)
		if userID == -1 {
			http.NotFound(w, r)
		}

		parking, err := parkingGetter.GetParkingByID(parkingID, userID)
		if err != nil {
			log.Error("error while getting Parking from DB", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		if parking == nil {
			http.NotFound(w, r)
			return
		}

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
		return 0, fmt.Errorf("не указан id парковки")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("id парковки %v не может быть преобразовано в число", idStr)
	}
	return id, err
}
