package parking

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	customValidator "github.com/PIRSON21/parking/internal/lib/validator"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=ParkingSetter
type ParkingSetter interface {
	AddParking(*models.Parking) error
	AddCellsForParking(*models.Parking, []*models.ParkingCellStruct) error
}

// AddParkingHandler создает парковку и добавляет в БД.
//
//goland:noinspection ALL
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

			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while decoding JSON: %w", op, err))

			return
		}

		valid := customValidator.CreateNewValidator()

		// валидируем данные
		if err = valid.Struct(&parking); err != nil {
			validateErr := err.(validator.ValidationErrors)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// добавляем данные в БД
		err = storage.AddParking(&parking)
		if err != nil {
			log.Error("error while adding Parking to DB", slog.String("err", err.Error()))

			resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while saving Parking: %w", op, err))

			return
		}

		var cellStruct []*models.ParkingCellStruct
		if parking.Cells != nil {
			var errs []error
			cellStruct, errs = createParkingCells(&parking)
			if errs != nil {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ListError("cells", errs))

				return
			}
		}

		if cellStruct != nil {
			err = storage.AddCellsForParking(&parking, cellStruct)
			if err != nil {
				log.Error("error while adding cells to DB", slog.String("err", err.Error()))

				resp.ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while adding cells to DB: %w", op, err))

				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// createParkingCells проверяет клетки парковки на соответствие требованиям.
// Возвращает список всех найденных ошибок
func createParkingCells(parking *models.Parking) ([]*models.ParkingCellStruct, []error) {
	var errors []error
	var cellStruct []*models.ParkingCellStruct

	if len(parking.Cells) != parking.Height {
		errors = append(errors, fmt.Errorf("ширина парковки не соответствует ширине топологии: %d", parking.Height))
	}

	for i, width := range parking.Cells {
		if len(width) != parking.Width {
			errors = append(errors, fmt.Errorf("длина строки %d не соответствует длине топологии: %d", i, parking.Width))
		}

		for j, cell := range width {
			if !cell.IsParkingCell() {
				errors = append(errors, fmt.Errorf("клетка (%d,%d) недействительна: '%s'", j, i, cell))
			}
			if !cell.IsRoad() {
				cellStruct = append(cellStruct, &models.ParkingCellStruct{X: j, Y: i, CellType: cell})
			}
		}
	}

	if len(errors) != 0 {
		return nil, errors
	}

	return cellStruct, nil
}
