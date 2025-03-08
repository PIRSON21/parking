package parking

import (
	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=ParkingGetter
type ParkingGetter interface {
	GetParkingsList(search string) ([]*models.Parking, error)
}

// AllParkingsHandler предоставляет обработчик получения списка парковок.
func AllParkingsHandler(log *slog.Logger, parkingGetter ParkingGetter, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handler.AllParkingsHandler"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		query := r.URL.Query().Get("search")

		parkings, err := parkingGetter.GetParkingsList(query)
		if err != nil {
			log.Error("error while getting from DB",
				slog.String("err", err.Error()))

			errorHandler(w, r, cfg, err)

			return
		}

		if len(parkings) == 0 {
			render.JSON(w, r, "")

			return
		}

		if err = render.RenderList(w, r, resp.NewParkingListRender(parkings)); err != nil {
			errorHandler(w, r, cfg, err)

			return
		}
	}
}

// errorHandler обрабатывает серверную ошибку (не клиентскую).
// Если приложение находится не в проде, выведет ошибку пользователю.
// Иначе, выведет стандартное сообщение "Internal Server Error".
func errorHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, err error) {
	if cfg.Environment != "prod" {
		renderError(w, r, err)
	} else {
		internalError(w)
	}
}

// internalError возвращает ошибку сервера без дополнительной информации для пользователя.
func internalError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// renderError предоставляет текст ошибки пользователя. Используется в версии для разработки.
func renderError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, resp.Error(err.Error()))
}
