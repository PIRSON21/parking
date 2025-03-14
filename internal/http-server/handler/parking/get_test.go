package parking_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
	"github.com/PIRSON21/parking/internal/lib/logger/handlers/slogdiscard"
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
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	urlAllParkings = "/parking"
)

func TestAllParkingsHandler_Success(t *testing.T) {
	t.Parallel()
	// Создаем мок для интерфейса ParkingGetter
	mockParkingGetter := new(mocks.ParkingGetter)
	mockParkingGetter.On("GetParkingsList", "").
		Return([]*models.Parking{
			{
				ID:      1,
				Name:    "1:Центр",
				Address: "ул. Ленина, 10",
				Width:   20,
				Height:  50,
			},
			{
				ID:      2,
				Name:    "1:ТЦ",
				Address: "ул. Ленина, 20",
				Width:   18,
				Height:  40,
			},
		}, nil)

	// Создаем HTTP-запрос
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlAllParkings, nil)
	assert.NoError(t, err, "")

	// Создаем ResponseRecorder (записывает ответ)
	rr := httptest.NewRecorder()

	// Создаем конфиг и логгер
	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{
		Environment: envLocal,
	}

	// Вызываем обработчик
	parking.AllParkingsHandler(log, mockParkingGetter, cfg).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()

	expectedJSON := `[{"id":1,"name":"1:Центр","address":"ул. Ленина, 10","height":50,"width":20,"url":"/parking/1"},{"id":2,"name":"1:ТЦ","address":"ул. Ленина, 20","height":40,"width":18,"url":"/parking/2"}]`
	assert.JSONEq(t, expectedJSON, body)
}

func TestAllParkingsHandler_SearchSuccess(t *testing.T) {
	t.Parallel()

	parkingGetterMock := new(mocks.ParkingGetter)
	parkingGetterMock.On("GetParkingsList", "центр абоба").
		Return([]*models.Parking{
			{
				ID:      1,
				Name:    "1:Центр",
				Address: "ул. Ленина, 10",
				Width:   20,
				Height:  50,
			},
			{
				ID:      2,
				Name:    "1:Центр",
				Address: "ул. Ленина, 20",
				Width:   18,
				Height:  40,
			},
		}, nil).
		Once()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/?search=центр абоба", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AllParkingsHandler(log, parkingGetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()

	expectedJSON := `[{"id":1,"name":"1:Центр","address":"ул. Ленина, 10","height":50,"width":20,"url":"/parking/1"},{"id":2,"name":"1:Центр","address":"ул. Ленина, 20","height":40,"width":18,"url":"/parking/2"}]`

	assert.JSONEq(t, expectedJSON, body)
}

func TestAllParkingsHandler_EmptyList(t *testing.T) {
	t.Parallel()
	// Создаем мок
	parkingGetterMock := new(mocks.ParkingGetter)
	parkingGetterMock.On("GetParkingsList", "Salam Aleykum").
		Return([]*models.Parking{}, nil)

	// Создаем request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/?search=Salam Aleykum", nil)
	require.NoError(t, err)

	// Создаем конфиг и логгер
	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	// Создаем тестовый сервер
	rr := httptest.NewRecorder()

	parking.AllParkingsHandler(log, parkingGetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()

	expectedJSON := `""`

	assert.JSONEq(t, expectedJSON, body)
}

func TestAllParkingsHandler_ErrorDev(t *testing.T) {
	t.Parallel()
	// Создаем mock
	parkingGetterMock := new(mocks.ParkingGetter)
	parkingGetterMock.On("GetParkingsList", "").
		Return(nil, errors.New("error test"))

	// Создаем запрос
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlAllParkings, nil)
	require.NoError(t, err)

	// Создаем обработчика
	rr := httptest.NewRecorder()

	// Создаем cfg и log
	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	// Запускаем обработчик
	parking.AllParkingsHandler(log, parkingGetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)

	body := rr.Body.String()

	expectedJSON := `{"error": "error test"}`
	assert.JSONEq(t, expectedJSON, body)
}

func TestAllParkingsHandler_ErrorProd(t *testing.T) {
	t.Parallel()

	parkingGetterMock := new(mocks.ParkingGetter)
	parkingGetterMock.On("GetParkingsList", "").
		Return(nil, errors.New("error test"))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlAllParkings, nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envProd}

	parking.AllParkingsHandler(log, parkingGetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)

	body := rr.Body.String()

	expectedBody := fmt.Sprintln("Internal Server Error")
	assert.Equal(t, expectedBody, body)

}

const (
	expectedJSONWithoutCells = `{"id":%d,"name":%q,"address":%q,"width":%d,"height":%d}`
	expectedJSONWithCells    = `{"id":%d,"name":%q,"address":%q,"width":%d,"height":%d,"cells":%s}`
	expectedError            = `{"error":%q}`

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
			Environment:     envProd,
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
			Environment:     envProd,
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
			cfg := &config.Config{Environment: envLocal}
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
				expectedResponse = fmt.Sprintf(expectedError, tc.GetParkingError.Error())
			}

			if tc.GetCellsError != nil {
				expectedResponse = fmt.Sprintf(expectedError, tc.GetCellsError.Error())
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
