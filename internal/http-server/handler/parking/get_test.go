package parking_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
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
			Name:   "Success with one parking",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:      1,
					Name:    "1: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
				},
			},
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: mustMarshalResponse([]resp.ParkingResponse{
				{
					ID:      1,
					Name:    "1: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
					URL:     "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with some parkings",
			Search: "",
			ParkingsList: []*models.Parking{
				{
					ID:      1,
					Name:    "1: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
				},
				{
					ID:      2,
					Name:    "2: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   4,
					Height:  4,
				},
			},
			GetParkingsError: nil,
			RequestURL:       urlAllParkings,
			ResponseCode:     http.StatusOK,
			ResponseBody: mustMarshalResponse([]resp.ParkingResponse{
				{
					ID:      1,
					Name:    "1: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
					URL:     "/parking/1",
				},
				{
					ID:      2,
					Name:    "2: Центр",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   4,
					Height:  4,
					URL:     "/parking/2",
				},
			}),
			JSON: true,
		},
		{
			Name:   "Success with search",
			Search: "aboba",
			ParkingsList: []*models.Parking{
				{
					ID:      1,
					Name:    "1: aboba",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
				},
			},
			GetParkingsError: nil,
			RequestURL:       fmt.Sprint(urlAllParkings + "?search=aboba"),
			ResponseCode:     http.StatusOK,
			ResponseBody: mustMarshalResponse([]resp.ParkingResponse{
				{
					ID:      1,
					Name:    "1: aboba",
					Address: "ул. Пушкина, д. Колотушкина",
					Width:   5,
					Height:  5,
					URL:     "/parking/1",
				},
			}),
			JSON: true,
		},
		{
			Name:             "Success empty list",
			ParkingsList:     nil,
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
			parkingGetterMock.On("GetParkingsList", tc.Search).
				Return(tc.ParkingsList, tc.GetParkingsError).
				Once()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.RequestURL, nil)
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
		// WithCells - ожидать JSON ответ с клетками или без
		WithCells bool
		// JSON - ожидать JSON ответ или нет
		// (при серверных ошибках на проде и при 404 ответ не JSON)
		JSON bool
		// Environment - значение из cfg. Для проверки ответов на проде и деве
		Environment string
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
			WithCells:       true,
			JSON:            true,
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
			WithCells:       false,
			JSON:            true,
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
			WithCells:       false,
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
			WithCells:       false,
			Environment:     test.EnvProd,
			JSON:            false,
		},
		{
			Name:            "Success not found",
			Parking:         nil,
			ResponseCode:    http.StatusNotFound,
			GetParkingError: sql.ErrNoRows,
			GetCellsError:   nil,
			WithCells:       false,
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
			WithCells:       false,
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
			WithCells:       false,
			Environment:     test.EnvProd,
			JSON:            false,
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

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, requestURL, nil)
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

			var expectedResponse string

			if tc.GetParkingError == nil && tc.GetCellsError == nil {
				if tc.WithCells {
					fmt.Println(tc.Parking.Cells)
					cells, err := json.Marshal(tc.Parking.Cells)
					require.NoError(t, err)
					expectedResponse = fmt.Sprintf(expectedJSONWithCells, tc.Parking.ID, tc.Parking.Name, tc.Parking.Address, tc.Parking.Width, tc.Parking.Height, cells)
				} else {
					expectedResponse = fmt.Sprintf(expectedJSONWithoutCells, tc.Parking.ID, tc.Parking.Name, tc.Parking.Address, tc.Parking.Width, tc.Parking.Height)
				}
			}

			if tc.GetParkingError != nil {
				expectedResponse = fmt.Sprintf(test.ExpectedError, tc.GetParkingError.Error())
			}

			if tc.GetCellsError != nil {
				expectedResponse = fmt.Sprintf(test.ExpectedError, tc.GetCellsError.Error())
			}

			if tc.JSON {
				assert.JSONEq(t, expectedResponse, body)

				return
			} else {
				if tc.ResponseCode == http.StatusInternalServerError {
					assert.Equal(t, "Internal Server Error\n", body)

					return
				} else if tc.ResponseCode == http.StatusNotFound {
					assert.Equal(t, "404 page not found\n", body)

					return
				}
			}

			assert.Fail(t, "для этого кейса не предусмотрен тест")
		})
	}
}

func mustMarshalResponse(v interface{}) string {
	var res []byte
	res, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(res)
}
