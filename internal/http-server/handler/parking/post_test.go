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
	required         = "Не указано поле"
	min              = "Минимальная длина поля %d"
	max              = "Максимальная длина поля %d"
	lte              = "Значение не может быть больше %d"
	gte              = "Значение не может быть меньше %d"
	cellsWidthWrong  = `"длина строки %d не соответствует длине топологии: %d"`
	cellsHeightWrong = `"ширина парковки не соответствует ширине топологии: %d"`
	cellsWrongCell   = `"клетка (%d,%d) недействительна: '%s'"`

	urlAddParking = "/add_parking"

	expectedErrors           = `{"error":[%s]}`
	expectedValidationError  = `{%q:%q}`
	expectedValidationErrors = `{%q:[%s]}`

	internalServerErrorMessage = "Internal Server Error\n"
)

func TestAddParkingHandler(t *testing.T) {
	cases := []struct {
		Name                    string
		RequestBody             []byte
		AddParkingError         error
		AddCellsForParkingError error
		ResponseCode            int
		ExpectedResponse        string
		JSON                    bool
		Environment             string
	}{
		{
			Name:                    "Wrong JSON format on local",
			RequestBody:             []byte(`{"wrong":"json"`),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        fmt.Sprintf(expectedError, "http-server.handler.parking.AddParkingHandler: error while decoding JSON: unexpected EOF"),
			JSON:                    true,
		},
		{
			Name:                    "Wrong JSON format on prod",
			RequestBody:             []byte(`{"wrong":"json"`),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        internalServerErrorMessage,
			JSON:                    false,
			Environment:             envProd,
		},
		{
			Name: "Success without cells",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusNoContent,
			ExpectedResponse:        "",
			JSON:                    false,
		},
		{
			Name: "Success with cells",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I"},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "."},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusNoContent,
			ExpectedResponse:        "",
			JSON:                    false,
		},
		{
			Name: "No name",
			RequestBody: mustMarshal(models.Parking{
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "name", required),
			JSON:                    true,
		},
		{
			Name: "No address",
			RequestBody: mustMarshal(models.Parking{
				Name:   "1: Центр",
				Width:  5,
				Height: 5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "address", required),
			JSON:                    true,
		},
		{
			Name: "No width",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "width", required),
			JSON:                    true,
		},
		{
			Name: "No height",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "height", required),
			JSON:                    true,
		},
		{
			Name: "Name length under min",
			RequestBody: mustMarshal(models.Parking{
				Name:    "a",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "name", fmt.Sprintf(min, 3)),
			JSON:                    true,
		},
		{
			Name: "Name length upper max",
			RequestBody: mustMarshal(models.Parking{
				Name:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "name", fmt.Sprintf(max, 10)),
			JSON:                    true,
		},
		{
			Name: "Address length under min",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "a",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "address", fmt.Sprintf(min, 10)),
			JSON:                    true,
		},
		{
			Name: "Address length upper max",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкинаaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "address", fmt.Sprintf(max, 30)),
			JSON:                    true,
		},
		{
			Name: "Width value under min",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   1,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "width", fmt.Sprintf(gte, 4)),
			JSON:                    true,
		},
		{
			Name: "Width value upper max",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   10,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "width", fmt.Sprintf(lte, 6)),
			JSON:                    true,
		},
		{
			Name: "Height value under min",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  1,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "height", fmt.Sprintf(gte, 4)),
			JSON:                    true,
		},
		{
			Name: "Height value upper max",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  10,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(expectedValidationError, "height", fmt.Sprintf(lte, 6)),
			JSON:                    true,
		},
		{
			Name: "Wrong parking width",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I", "."},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "."},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse: fmt.Sprintf(expectedValidationErrors, "cells",
				fmt.Sprintf(cellsWidthWrong, 0, 5)),
			JSON: true,
		},
		{
			Name: "Wrong parking height",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I"},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "."},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse: fmt.Sprintf(expectedValidationErrors, "cells",
				fmt.Sprintf(cellsHeightWrong, 5)),
			JSON: true,
		},
		{
			Name: "Wrong parking cell",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I"},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "H"},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse: fmt.Sprintf(expectedValidationErrors, "cells",
				fmt.Sprintf(cellsWrongCell, 4, 2, "H")),
			JSON: true,
		},
		{
			Name: "Internal addParking error on dev",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         fmt.Errorf("test parking error"),
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        fmt.Sprintf(expectedError, "http-server.handler.parking.AddParkingHandler: error while saving Parking: test parking error"),
			JSON:                    true,
			Environment:             envLocal,
		},
		{
			Name: "Internal addParking error on prod",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         fmt.Errorf("test parking error"),
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        internalServerErrorMessage,
			JSON:                    false,
			Environment:             envProd,
		},
		{
			Name: "Internal addCellsForParking error on dev",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I"},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "."},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: fmt.Errorf("test parking error"),
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        fmt.Sprintf(expectedError, "http-server.handler.parking.AddParkingHandler: error while adding cells to DB: test parking error"),
			JSON:                    true,
			Environment:             envLocal,
		},
		{
			Name: "Internal addCellsForParking error on prod",
			RequestBody: mustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
				Cells: [][]models.ParkingCell{
					{".", ".", ".", ".", "I"},
					{".", "P", "P", "P", "."},
					{".", "D", "D", ".", "."},
					{".", ".", ".", ".", "."},
					{"O", ".", ".", "P", "P"},
				},
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: fmt.Errorf("test parking error"),
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        internalServerErrorMessage,
			JSON:                    false,
			Environment:             envProd,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			parkingSetterMock := mocks.NewParkingSetter(t)
			parkingSetterMock.On("AddParking", mock.AnythingOfType("*models.Parking")).
				Return(tc.AddParkingError).
				Maybe()

			parkingSetterMock.On("AddCellsForParking",
				mock.AnythingOfType("*models.Parking"),
				mock.AnythingOfType("[]*models.ParkingCellStruct")).
				Return(tc.AddCellsForParkingError).
				Maybe()

			reqBody := bytes.NewReader(tc.RequestBody)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: envLocal}

			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			parking.AddParkingHandler(log, parkingSetterMock, cfg).ServeHTTP(rr, req)
			require.Equal(t, tc.ResponseCode, rr.Code)

			body := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ExpectedResponse, body)

				return
			} else {
				assert.Equal(t, tc.ExpectedResponse, body)

				return
			}

			assert.Fail(t, "не все проверки прописаны")
		})
	}

}

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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, body)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlAddParking, reqBody)
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

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
