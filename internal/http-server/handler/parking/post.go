package parking

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	customValidator "github.com/PIRSON21/parking/internal/lib/validator"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/xerrors"
)

var invalidParkingIndex = errors.New("invalid parkingID syntax")

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=ParkingSetter
type ParkingSetter interface {
	AddParking(*models.Parking) error
	DeleteParking(int) error
	UpdateParking(*ParkingPatch, []*models.ParkingCellStruct) (*models.Parking, error)
}

// AddParkingHandler создает парковку и добавляет в БД.
//
//goland:noinspection t
func AddParkingHandler(log *slog.Logger, storage ParkingSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.parking.AddParkingHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		defer r.Body.Close()

		// считываем данные о парковке
		var parking models.Parking
		err := render.DecodeJSON(r.Body, &parking)
		if err != nil {
			log.Error("error while decoding JSON", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while decoding JSON: %w", op, err))
			return
		}
		log.Debug("parking from request", slog.Any("parking", parking))

		valid := customValidator.CreateNewValidator()

		// валидируем данные
		if err = valid.Struct(&parking); err != nil {
			log.Error("validation error", slog.String("err", err.Error()))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		log.Debug("parking validated", slog.Any("parking", parking))

		if parking.Cells != nil {
			log.Debug("validating parking cells")
			errs := customValidator.ValidateParkingCells(&parking)
			if errs != nil {
				log.Error("error while validating parking cells", slog.Any("errors", errs))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ListError("cells", errs))
				return
			}
		}

		log.Debug("adding parking to DB")
		// добавляем данные в БД
		err = storage.AddParking(&parking)
		if err != nil {
			log.Error("error while adding Parking to DB", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, xerrors.Errorf("%s: error while saving Parking: %w", op, err))
			return
		}

		log.Debug("parking added to DB", slog.Int("parkingID", parking.ID), slog.String("name", parking.Name))
		w.WriteHeader(http.StatusCreated)
	}
}

// DeleteParkingHandler удаляет парковку.
func DeleteParkingHandler(log *slog.Logger, db ParkingSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.parking.post.DeleteParkingHandler"

		log = log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
			slog.String("op", op),
		)

		parkingID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("error while getting parkingID", slog.String("err", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError("invalid parkingID syntax"))
			return
		}
		log.Debug("parkingID from url", slog.Int("parkingID", parkingID))

		err = db.DeleteParking(parkingID)
		if err != nil {
			log.Error("error while deleting parking", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}

		log.Debug("parking deleted", slog.Int("parkingID", parkingID))
		w.WriteHeader(http.StatusNoContent)
	}
}

type ParkingPatch struct {
	ID          int                    `json:"id,omitempty"`
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=3,max=10"`
	Address     *string                `json:"address,omitempty" validate:"omitempty,min=10,max=30"`
	Width       *int                   `json:"width,omitempty" validate:"omitempty,gte=4,lte=6"`
	Height      *int                   `json:"height,omitempty" validate:"omitempty,gte=4,lte=6"`
	DayTariff   *int                   `json:"day_tariff,omitempty" validate:"omitempty,gte=0,lte=1000"`
	NightTariff *int                   `json:"night_tariff,omitempty" validate:"omitempty,gte=0,lte=1000"`
	Cells       [][]models.ParkingCell `json:"cells,omitempty"`
	Manager     *models.Manager        `json:"manager,omitempty"`
}

// UpdateParkingHandler обновляет данные о парковке.
//
//goland:noinspection t
func UpdateParkingHandler(log *slog.Logger, db ParkingSetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.parking.post.UpdateParkingHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		parkingID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("error while parsing parkingID", slog.String("err", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.UnknownError(invalidParkingIndex.Error()))
			return
		}
		log.Debug("parkingID from url", slog.Int("parkingID", parkingID))

		var parkingUpdates ParkingPatch
		err = render.DecodeJSON(r.Body, &parkingUpdates)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			log.Error("error while decoding JSON", slog.String("err", err.Error()))
			render.JSON(w, r, resp.UnknownError(fmt.Sprintf("error while decoding JSON: %s", err.Error())))
			return
		}
		log.Debug("parkingUpdates from request", slog.Any("parkingUpdates", parkingUpdates))
		parkingUpdates.ID = parkingID

		valid := customValidator.CreateNewValidator()
		if err = valid.Struct(&parkingUpdates); err != nil {
			var valErr validator.ValidationErrors
			if errors.As(err, &valErr) {
				log.Error("validation error", slog.String("err", err.Error()))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ValidationError(valErr))
				return
			}
			log.Error("error while validating parking updates", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}

		var cellStruct []*models.ParkingCellStruct
		if parkingUpdates.Cells != nil {
			var errs []error
			parking := models.Parking{
				ID:     parkingUpdates.ID,
				Width:  *parkingUpdates.Width,
				Height: *parkingUpdates.Height,
				Cells:  parkingUpdates.Cells,
			}
			errs = customValidator.ValidateParkingCells(&parking)
			if errs != nil {
				log.Error("error while validating parking cells", slog.Any("errors", errs))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ListError("cells", errs))
				return
			}
		}

		log.Debug("updating parkings")
		parking, err := db.UpdateParking(&parkingUpdates, cellStruct)
		if err != nil {
			log.Error("error while updating parking", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		log.Debug("parking updated", slog.Int("parkingID", parking.ID), slog.String("name", parking.Name))

		render.JSON(w, r, parking)
	}
}
