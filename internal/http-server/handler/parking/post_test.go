package parking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
	"github.com/PIRSON21/parking/internal/lib/logger/handlers/slogdiscard"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	required = "Не указано поле"
	min      = "Минимальная длина поля %d"
	max      = "Максимальная длина поля %d"
	lte      = "Значение не может быть больше %d"
	gte      = "Значение не может быть меньше %d"
)

func TestAddParkingHandler_Success(t *testing.T) {
	t.Parallel()

	// добавляем моки
	parkingSetterMock := mocks.NewParkingSetter(t)
	parkingSetterMock.On("AddParking", mock.AnythingOfType("*models.Parking")).
		Return(nil)

	// добавляем данные для проверки
	p := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, 10",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(p)
	body := bytes.NewReader(jsonBody)

	// создаем клиентский запрос
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", body)
	require.NoError(t, err)

	// создаем аля сервер
	rr := httptest.NewRecorder()

	// создаем log, cfg
	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	bodyRes := rr.Body.String()

	assert.Equal(t, "", bodyRes)
}

func TestAddParkingHandler_NoName(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Address: "ул. Ленина, д. 10",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := required

	expectedJSON := `{"name":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_NoAddress(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:   "Парковка",
		Width:  5,
		Height: 5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := required

	expectedJSON := `{"address":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_NoWidth(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := required

	expectedJSON := `{"width":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_NoHeight(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Width:   5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := required

	expectedJSON := `{"height":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MinName(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "1",
		Address: "ул. Ленина, д. 10",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(min, 3)

	expectedJSON := `{"name":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MaxName(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "111111111111111111111111111111",
		Address: "ул. Ленина, д. 10",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(max, 10)

	expectedJSON := `{"name":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MinAddress(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул.",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(min, 10)

	expectedJSON := `{"address":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MaxAddress(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул.111111111111111111111111111111111111111111111111111",
		Width:   5,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(max, 30)

	expectedJSON := `{"address":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MinWidth(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Width:   2,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(gte, 4)

	expectedJSON := `{"width":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MaxWidth(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Width:   7,
		Height:  5,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(lte, 6)

	expectedJSON := `{"width":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MinHeight(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Width:   5,
		Height:  2,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(gte, 4)

	expectedJSON := `{"height":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}

func TestAddParkingHandler_MaxHeigth(t *testing.T) {
	t.Parallel()

	parkingSetterMock := mocks.NewParkingSetter(t)

	parkingInput := models.Parking{
		Name:    "Парковка",
		Address: "ул. Ленина, д. 10",
		Width:   5,
		Height:  7,
	}
	jsonBody, _ := json.Marshal(parkingInput)
	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/add_parking", reqBody)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	log := slogdiscard.NewDiscardLogger()
	cfg := &config.Config{Environment: envLocal}

	parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	respBody := strings.TrimSpace(rr.Body.String())

	errMessage := fmt.Sprintf(lte, 6)

	expectedJSON := `{"height":"` + errMessage + `"}`

	assert.Equal(t, expectedJSON, respBody)
}
