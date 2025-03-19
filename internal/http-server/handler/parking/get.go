package parking

import (
	"database/sql"
	"errors"
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
	GetAdminParkings(search string) ([]*models.Parking, error)
	GetManagerParkings(userID int, search string) ([]*models.Parking, error)
	GetParkingByID(parkingID int) (*models.Parking, error)
	GetParkingCells(parking *models.Parking) ([][]models.ParkingCell, error)
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
func GetParkingHandler(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.Parking.GetParkingHandler"

		log := log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError("не указан id парковки"))

			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(fmt.Sprintf("id парковки %v не может быть преобразовано в число", idStr)))

			return
		}

		parking, err := parkingGetter.GetParkingByID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)

				return
			}

			log.Error("error while getting Parking from DB", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, err)

			return
		}

		parking.Cells, err = parkingGetter.GetParkingCells(parking)
		if err != nil {
			log.Error("error while getting cells from DB", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, err)

			return
		}

		render.JSON(w, r, parking)
	}
}
