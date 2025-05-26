package ws

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PIRSON21/parking/internal/config"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	custom_validator "github.com/PIRSON21/parking/internal/lib/validator"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/PIRSON21/parking/internal/simulation"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(log *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.With(
			slog.String("reqID", middleware.GetReqID(r.Context())),
		)

		// upgrade rest request to websocket connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("error while upgrading webSocket conn", slog.String("err", err.Error()))
			resp.ErrorHandler(w, r, cfg, err)
			return
		}
		defer conn.Close()

		// создаем новый клиент
		client := NewClient(conn)

		var initParams struct {
			Parking           *models.Parking               `json:"parking" validate:"required"`
			ArrivalConfig     *simulation.ArrivalConfig     `json:"arrival_config" validate:"required"`
			ParkingTimeConfig *simulation.ParkingTimeConfig `json:"parking_time_config" validate:"required"`
			StartTime         int64                         `json:"start_time" validate:"required"`
		}

		err = conn.ReadJSON(&initParams)
		if err != nil {
			log.Error("error while reading params", slog.String("err", err.Error()))
			conn.WriteJSON(resp.UnknownError("error while reading params"))
			return
		}
		log.Debug("params from client", slog.Any("params", initParams))

		valid := custom_validator.CreateNewValidator()
		// добавление кастомной валидации для параметров моделирования
		valid.RegisterStructValidation(custom_validator.ArrivalConfigStructLevelValidation, simulation.ArrivalConfig{})
		valid.RegisterStructValidation(custom_validator.ParkingTimeConfigStructLevelValidation, simulation.ParkingTimeConfig{})
		if err := valid.Struct(&initParams); err != nil {
			log.Error("validation error", slog.String("err", err.Error()))
			validErr := err.(validator.ValidationErrors)
			conn.WriteJSON(resp.RecursiveValidationError(validErr))
			return
		}
		log.Debug("params validation passed", slog.Any("params", initParams))

		// создаем сессию клиента
		session := simulation.NewSession(
			client, initParams.Parking, time.Unix(initParams.StartTime, 0),
			initParams.ArrivalConfig, initParams.ParkingTimeConfig, log,
		)
		log.Debug("session created", slog.Any("session", session))

		go client.WriteLoop(log)

		go client.ReadLoop(log, readFunc(session, client))

		client.Send([]byte("ok"))

		<-client.Done
		session.Stop()
	}
}

func readFunc(session *simulation.Session, client *Client) func(msg []byte) {
	return func(msg []byte) {
		switch string(msg) {
		case "start":
			session.Start()
		case "pause":
			session.Pause()
		case "resume":
			session.Resume()
		case "stop":
			session.Stop()
			client.Stop()
		default:
			if str := string(msg); strings.HasPrefix(str, "park") {
				go session.CheckPark(str)
			}
		}
	}
}
