package parking

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=ParkingSetter
type ParkingSetter interface {
	AddParking(*models.Parking) error
}

// AddParkingHandler создает парковку и добавляет в БД.
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

			ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while decoding JSON: %w", op, err))

			return
		}

		valid := CreateNewValidator()

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
			log.Error("error while adding parking to DB", slog.String("err", err.Error()))

			ErrorHandler(w, r, cfg, fmt.Errorf("%s: error while saving parking: %w", op, err))

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func CreateNewValidator() *validator.Validate {
	valid := validator.New()

	valid.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return valid
}
