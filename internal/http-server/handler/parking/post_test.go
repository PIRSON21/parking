package parking_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking/mocks"
	"github.com/PIRSON21/parking/internal/lib/logger/handlers/slogdiscard"
	"github.com/PIRSON21/parking/internal/lib/test"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	cellsWidthWrong  = `"длина строки %d не соответствует длине топологии: %d"`
	cellsHeightWrong = `"ширина парковки не соответствует ширине топологии: %d"`
	cellsWrongCell   = `"клетка (%d,%d) недействительна: '%s'"`

	urlAddParking = "/add_parking"
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
			ExpectedResponse:        fmt.Sprintf(test.ExpectedError, "http-server.handler.parking.AddParkingHandler: error while decoding JSON: unexpected EOF"),
			JSON:                    true,
		},
		{
			Name:                    "Wrong JSON format on prod",
			RequestBody:             []byte(`{"wrong":"json"`),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        test.InternalServerErrorMessage,
			JSON:                    false,
			Environment:             test.EnvProd,
		},
		{
			Name: "Success without cells",
			RequestBody: test.MustMarshal(models.Parking{
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
			RequestBody: test.MustMarshal(models.Parking{
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
			RequestBody: test.MustMarshal(models.Parking{
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "name", test.Required),
			JSON:                    true,
		},
		{
			Name: "No address",
			RequestBody: test.MustMarshal(models.Parking{
				Name:   "1: Центр",
				Width:  5,
				Height: 5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "address", test.Required),
			JSON:                    true,
		},
		{
			Name: "No width",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "width", test.Required),
			JSON:                    true,
		},
		{
			Name: "No height",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "height", test.Required),
			JSON:                    true,
		},
		{
			Name: "Name length under min",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "a",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "name", fmt.Sprintf(test.Min, 3)),
			JSON:                    true,
		},
		{
			Name: "Name length upper max",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "name", fmt.Sprintf(test.Max, 10)),
			JSON:                    true,
		},
		{
			Name: "Address length under min",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "a",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "address", fmt.Sprintf(test.Min, 10)),
			JSON:                    true,
		},
		{
			Name: "Address length upper max",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкинаaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "address", fmt.Sprintf(test.Max, 30)),
			JSON:                    true,
		},
		{
			Name: "Width value under min",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   1,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "width", fmt.Sprintf(test.Gte, 4)),
			JSON:                    true,
		},
		{
			Name: "Width value upper max",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   10,
				Height:  5,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "width", fmt.Sprintf(test.Lte, 6)),
			JSON:                    true,
		},
		{
			Name: "Height value under min",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  1,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "height", fmt.Sprintf(test.Gte, 4)),
			JSON:                    true,
		},
		{
			Name: "Height value upper max",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  10,
			}),
			AddParkingError:         nil,
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusBadRequest,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedValidationError, "height", fmt.Sprintf(test.Lte, 6)),
			JSON:                    true,
		},
		{
			Name: "Wrong parking width",
			RequestBody: test.MustMarshal(models.Parking{
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
			ExpectedResponse: fmt.Sprintf(test.ExpectedValidationErrors, "cells",
				fmt.Sprintf(cellsWidthWrong, 0, 5)),
			JSON: true,
		},
		{
			Name: "Wrong parking height",
			RequestBody: test.MustMarshal(models.Parking{
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
			ExpectedResponse: fmt.Sprintf(test.ExpectedValidationErrors, "cells",
				fmt.Sprintf(cellsHeightWrong, 5)),
			JSON: true,
		},
		{
			Name: "Wrong parking cell",
			RequestBody: test.MustMarshal(models.Parking{
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
			ExpectedResponse: fmt.Sprintf(test.ExpectedValidationErrors, "cells",
				fmt.Sprintf(cellsWrongCell, 4, 2, "H")),
			JSON: true,
		},
		{
			Name: "Internal addParking error on dev",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         fmt.Errorf("test parking error"),
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        fmt.Sprintf(test.ExpectedError, "http-server.handler.parking.AddParkingHandler: error while saving Parking: test parking error"),
			JSON:                    true,
			Environment:             test.EnvLocal,
		},
		{
			Name: "Internal addParking error on prod",
			RequestBody: test.MustMarshal(models.Parking{
				Name:    "1: Центр",
				Address: "ул. Пушкина, д. Колотушкина",
				Width:   5,
				Height:  5,
			}),
			AddParkingError:         fmt.Errorf("test parking error"),
			AddCellsForParkingError: nil,
			ResponseCode:            http.StatusInternalServerError,
			ExpectedResponse:        test.InternalServerErrorMessage,
			JSON:                    false,
			Environment:             test.EnvProd,
		},
		{
			Name: "Internal addCellsForParking error on dev",
			RequestBody: test.MustMarshal(models.Parking{
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
			ExpectedResponse:        fmt.Sprintf(test.ExpectedError, "http-server.handler.parking.AddParkingHandler: error while adding cells to DB: test parking error"),
			JSON:                    true,
			Environment:             test.EnvLocal,
		},
		{
			Name: "Internal addCellsForParking error on prod",
			RequestBody: test.MustMarshal(models.Parking{
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
			ExpectedResponse:        test.InternalServerErrorMessage,
			JSON:                    false,
			Environment:             test.EnvProd,
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
			cfg := &config.Config{Environment: test.EnvLocal}

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
