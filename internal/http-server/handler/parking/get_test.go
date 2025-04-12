package parking_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
	authMiddleware "github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	"github.com/PIRSON21/parking/internal/lib/logger/handlers/slogdiscard"
	"github.com/PIRSON21/parking/internal/lib/test"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	urlAllParkings = "/parking"
)

func TestAllParkingsHandler(t *testing.T) {
	cases := []struct {
		Name             string
		UserID           int
		Search           string
		ParkingsList     []*models.Parking
		GetParkingsError error
		RequestURL       string
		ResponseCode     int
		ResponseBody     string
		JSON             bool
		Environment      string
	}{
		{
			Name:   "Success with one parking as admin",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with one parking as manager",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			UserID:           1,
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with some parkings as admin",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
				{
					ID:          2,
					Name:        "2: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       4,
					Height:      4,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
				{
					ID:          2,
					Name:        "2: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/2",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with some parkings as manager",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
				{
					ID:          2,
					Name:        "2: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       4,
					Height:      4,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			UserID:           1,
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
				{
					ID:          2,
					Name:        "2: Центр",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/2",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with search as admin",
			Search: "aboba",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: aboba",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			GetParkingsError: nil,
			RequestURL:       fmt.Sprint(urlAllParkings + "?search=aboba"),
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: aboba",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with search as manager",
			Search: "aboba",
			ParkingsList: []*models.Parking{
				{
					ID:          1,
					Name:        "1: aboba",
					Address:     "ул. Пушкина, д. Колотушкина",
					Width:       5,
					Height:      5,
					DayTariff:   5,
					NightTariff: 1,
				},
			},
			UserID:           1,
			GetParkingsError: nil,
			RequestURL:       fmt.Sprint(urlAllParkings + "?search=aboba"),
			ResponseCode:     http.StatusOK,
			ResponseBody: test.MustMarshalResponse([]resp.ParkingResponse{
				{
					ID:          1,
					Name:        "1: aboba",
					Address:     "ул. Пушкина, д. Колотушкина",
					DayTariff:   5,
					NightTariff: 1,
					URL:         "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:             "Success empty list as admin",
			ParkingsList:     nil,
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody:     "[]",
			JSON:             true,
		},
		{
			Name:             "Success empty list as manager",
			ParkingsList:     nil,
			UserID:           1,
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody:     "[]",
			JSON:             true,
		},
		{
			Name:             "Error while getting parks on dev",
			Search:           "",
			ParkingsList:     nil,
			GetParkingsError: fmt.Errorf("parking getter error"),
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusInternalServerError,
			ResponseBody:     fmt.Sprintf(test.ExpectedError, "parking getter error"),
			JSON:             true,
			Environment:      test.EnvLocal,
		},
		{
			Name:             "Error while getting parks on prod",
			Search:           "",
			ParkingsList:     nil,
			GetParkingsError: fmt.Errorf("parking getter error"),
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusInternalServerError,
			ResponseBody:     test.InternalServerErrorMessage,
			JSON:             false,
			Environment:      test.EnvProd,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			parkingGetterMock := mocks.NewParkingGetter(t)

			parkingGetterMock.On("GetAdminParkings", tc.Search).
				Return(tc.ParkingsList, tc.GetParkingsError).
				Maybe()

			parkingGetterMock.On("GetManagerParkings", tc.UserID, tc.Search).
				Return(tc.ParkingsList, tc.GetParkingsError).
				Maybe()

			newCtx := context.WithValue(context.Background(), authMiddleware.UserIDKey, tc.UserID)
			req, err := http.NewRequestWithContext(newCtx, http.MethodGet, tc.RequestURL, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: test.EnvLocal}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			parking.AllParkingsHandler(log, parkingGetterMock, cfg).ServeHTTP(rr, req)
			require.Equal(t, tc.ResponseCode, rr.Code)

			body := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, body)

				return
			} else {
				assert.Equal(t, tc.ResponseBody, body)

				return
			}

			assert.Fail(t, "прописаны не все проверки")
		})
	}
}

const (
	expectedJSONWithoutCells = `{"id":%d,"name":%q,"address":%q,"width":%d,"height":%d}`
	expectedJSONWithCells    = `{"id":%d,"name":%q,"address":%q,"width":%d,"height":%d,"cells":%s}`
	expectedJSONWith

	urlCurrentParking = "/parking/%d"
)

// TestGetParkingHandler проверяет запрос получения данных о парковке по ID
//
// Если изменится адрес или ответ, измените шаблон сверху
//
//goland:noinspection t
func TestGetParkingHandler(t *testing.T) {

	cases := []struct {
		// Name - название теста. Нужен только для отображения
		Name string
		// Parking - указатель на модель. От неё все зависит - ответы, запросы и т.д.
		// Может быть nil, тогда запрос к БД будет на id = 1, а в результате будет nil.
		Parking *models.Parking
		// ResponseCode - ожидаемый код ответа. Лучше подбирать через пакет http
		ResponseCode int
		// GetParkingError - ожидаемая ошибка от метода GetParkingByID. Может быть nil
		GetParkingError error
		// GetCellsError - ожидаемая ошибка от метода GetParkingCells. Может быть nil
		GetCellsError error
		// JSON - ожидать JSON ответ или нет
		// (при серверных ошибках на проде и при 404 ответ не JSON)
		JSON bool
		// Environment - значение из cfg. Для проверки ответов на проде и деве
		Environment string
		// UserID - id пользователя, который будет передан в контекст. 0 - админ, > 0 - менеджера
		UserID int
		// ResponseBody - тело ответа
		ResponseBody string
	}{
		{
			Name: "Success With Cells",
			Parking: &models.Parking{
				ID:      1,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells: [][]models.ParkingCell{
					{
						"P", "P", "P", "P",
					},
					{
						".", ".", ".", ".",
					},
					{
						"P", "P", ".", ".",
					},
					{
						"P", "O", "I", "P",
					},
				},
			},
			ResponseCode:    http.StatusOK,
			GetParkingError: nil,
			GetCellsError:   nil,
			ResponseBody: test.MustMarshalResponse(&models.Parking{
				ID:      1,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells: [][]models.ParkingCell{
					{
						"P", "P", "P", "P",
					},
					{
						".", ".", ".", ".",
					},
					{
						"P", "P", ".", ".",
					},
					{
						"P", "O", "I", "P",
					},
				},
			}),
			JSON: true,
		},
		{
			Name: "Success without cells",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			},
			ResponseCode:    http.StatusOK,
			GetParkingError: nil,
			GetCellsError:   nil,
			ResponseBody: test.MustMarshalResponse(&models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			}),
			JSON: true,
		},
		{
			Name: "Error while getting from DB on Dev",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			},
			ResponseCode:    http.StatusInternalServerError,
			GetParkingError: fmt.Errorf("db: error getting from DB"),
			GetCellsError:   nil,
			ResponseBody:    fmt.Sprintf(test.ExpectedError, "db: error getting from DB"),
			JSON:            true,
		},
		{
			Name: "Error while getting from DB on Prod",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			},
			ResponseCode:    http.StatusInternalServerError,
			GetParkingError: fmt.Errorf("db: error getting from DB"),
			GetCellsError:   nil,
			Environment:     test.EnvProd,
			ResponseBody:    test.InternalServerErrorMessage,
			JSON:            false,
		},
		{
			Name:            "Success not found",
			Parking:         nil,
			ResponseCode:    http.StatusNotFound,
			GetParkingError: sql.ErrNoRows,
			GetCellsError:   nil,
			ResponseBody:    test.NotFound,
			JSON:            false,
		},
		{
			Name: "Error while getting cells on Dev",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			},
			ResponseCode:    http.StatusInternalServerError,
			GetParkingError: nil,
			GetCellsError:   fmt.Errorf("cells: error while getting cells"),
			ResponseBody:    fmt.Sprintf(test.ExpectedError, "cells: error while getting cells"),
			JSON:            true,
		},
		{
			Name: "Error while getting cells on Prod",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			},
			ResponseCode:    http.StatusInternalServerError,
			GetParkingError: nil,
			GetCellsError:   fmt.Errorf("cells: error while getting cells"),
			Environment:     test.EnvProd,
			ResponseBody:    test.InternalServerErrorMessage,
			JSON:            false,
		},
		{
			Name: "Success get parking with manager to manager",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
				Manager: &models.Manager{ID: 1},
			},
			ResponseCode:    http.StatusOK,
			GetParkingError: nil,
			GetCellsError:   nil,
			JSON:            true,
			ResponseBody: test.MustMarshalResponse(&models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
			}),
			UserID: 1,
		},
		{
			Name: "Success get parking with manager to admin",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
				Manager: &models.Manager{ID: 1},
			},
			ResponseCode:    http.StatusOK,
			GetParkingError: nil,
			GetCellsError:   nil,
			JSON:            true,
			ResponseBody: test.MustMarshalResponse(&models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
				Manager: &models.Manager{ID: 1},
			}),
			UserID: 0,
		},
		{
			Name: "Success not allowed parking for manager",
			Parking: &models.Parking{
				ID:      2,
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   4,
				Height:  4,
				Cells:   nil,
				Manager: &models.Manager{ID: 2},
			},
			ResponseCode:    http.StatusNotFound,
			GetParkingError: nil,
			GetCellsError:   nil,
			JSON:            false,
			ResponseBody:    test.NotFound,
			UserID:          1,
		},
	}

	t.Parallel()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			parkingGetterMock := mocks.NewParkingGetter(t)
			if tc.Parking != nil {
				// Если нужны данные парковки
				parkingGetterMock.On("GetParkingByID", tc.Parking.ID).
					Return(&models.Parking{
						ID:      tc.Parking.ID,
						Name:    tc.Parking.Name,
						Address: tc.Parking.Address,
						Width:   tc.Parking.Width,
						Height:  tc.Parking.Height,
						Manager: tc.Parking.Manager,
					}, tc.GetParkingError).
					Once()
			} else {
				// Если парковка не нужна, задает структура только для отправки запроса по ID
				tc.Parking = &models.Parking{ID: 1}
				parkingGetterMock.On("GetParkingByID", tc.Parking.ID).
					Return(nil, tc.GetParkingError).
					Once()
			}

			parkingGetterMock.On("GetParkingCells", mock.AnythingOfType("*models.Parking")).
				Return(tc.Parking.Cells, tc.GetCellsError).
				Maybe()

			requestURL := fmt.Sprintf(urlCurrentParking, tc.Parking.ID)

			ctx := context.WithValue(context.Background(), authMiddleware.UserIDKey, tc.UserID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
			require.NoError(t, err)

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: test.EnvLocal}
			if tc.Environment != "" {
				// Если указан специфичный Environment
				cfg.Environment = tc.Environment
			}

			router := chi.NewRouter()
			router.Use(middleware.URLFormat)
			router.Get("/parking/{id}", parking.GetParkingHandler(log, parkingGetterMock, cfg))

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)
			require.Equal(t, tc.ResponseCode, rr.Code)

			body := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, body)
				return
			} else {
				assert.Equal(t, tc.ResponseBody, body)
				return
			}

			assert.Fail(t, "для этого кейса не предусмотрен тест")
		})
	}
}
