package parking_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
	"github.com/PIRSON21/parking/internal/lib/logger/handlers/slogdiscard"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
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

	expectedJSON := `[{"id":1,"name":"1:Центр","address":"ул. Ленина, 10","height":50,"width":20},{"id":2,"name":"1:ТЦ","address":"ул. Ленина, 20","height":40,"width":18}]`
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

	expectedJSON := `[{"id":1,"name":"1:Центр","address":"ул. Ленина, 10","height":50,"width":20},{"id":2,"name":"1:Центр","address":"ул. Ленина, 20","height":40,"width":18}]`

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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
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
