package user_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/user"
	"github.com/PIRSON21/parking/internal/http-server/handler/user/mocks"
	resp "github.com/PIRSON21/parking/internal/lib/api/response"
	customErr "github.com/PIRSON21/parking/internal/lib/errors"
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
	"strconv"
	"testing"
)

const (
	loginURL = "/login"
)

func TestLoginHandler(t *testing.T) {
	cases := []struct {
		Name                     string
		RequestBody              []byte
		ManagerID                int
		AuthenticateManagerError error
		SetSessionIDError        error
		Environment              string
		ResponseCode             int
		JSON                     bool
		ResponseBody             string
	}{
		{
			Name: "Success with manager",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
			}),
			ManagerID:                1,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusAccepted,
			JSON:                     false,
			ResponseBody:             "",
		},
		{
			Name: "Success with admin",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "admin",
				Password: "admin",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusAccepted,
			JSON:                     false,
			ResponseBody:             "",
		},
		{
			Name: "Wrong account",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "wrong",
				Password: "wrong",
			}),
			ManagerID:                0,
			AuthenticateManagerError: customErr.ErrUnauthorized,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusNotFound,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedError, "неправильный логин или пароль"),
		},
		{
			Name: "Login smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "a",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusBadRequest,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedValidationError, "login", fmt.Sprintf(test.Min, 4)),
		},
		{
			Name: "Login bigger max",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusBadRequest,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedValidationError, "login", fmt.Sprintf(test.Max, 8)),
		},
		{
			Name: "Password smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "a",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusBadRequest,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedValidationError, "password", fmt.Sprintf(test.Min, 4)),
		},
		{
			Name: "Password bigger max",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusBadRequest,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedValidationError, "password", fmt.Sprintf(test.Max, 10)),
		},
		{
			Name: "Validation errors",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "a",
				Password: "a",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        nil,
			ResponseCode:             http.StatusBadRequest,
			JSON:                     true,
			ResponseBody: string(test.MustMarshal(map[string]string{
				"login":    fmt.Sprintf(test.Min, 4),
				"password": fmt.Sprintf(test.Min, 4),
			})),
		},
		{
			Name: "Internal AuthenticateManager error on prod",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: fmt.Errorf("test authenticateManager error"),
			SetSessionIDError:        nil,
			Environment:              test.EnvLocal,
			ResponseCode:             http.StatusInternalServerError,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedError, "http-server.handler.user.LoginHandler: error while getting manager info: test authenticateManager error"),
		},
		{
			Name: "Internal AuthenticateManager error on prod",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: fmt.Errorf("test authenticateManager error"),
			SetSessionIDError:        nil,
			Environment:              test.EnvProd,
			ResponseCode:             http.StatusInternalServerError,
			JSON:                     false,
			ResponseBody:             test.InternalServerErrorMessage,
		},
		{
			Name: "Internal SetSessionID error on prod",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        fmt.Errorf("test SetSessionID error"),
			Environment:              test.EnvLocal,
			ResponseCode:             http.StatusInternalServerError,
			JSON:                     true,
			ResponseBody:             fmt.Sprintf(test.ExpectedError, "http-server.handler.user.LoginHandler: error while returning session: http-server.handler.user.returnSessionID: error while setting session to DB: test SetSessionID error"),
		},
		{
			Name: "Internal SetSessionID error on prod",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
			}),
			ManagerID:                0,
			AuthenticateManagerError: nil,
			SetSessionIDError:        fmt.Errorf("test SetSessionID error"),
			Environment:              test.EnvProd,
			ResponseCode:             http.StatusInternalServerError,
			JSON:                     false,
			ResponseBody:             test.InternalServerErrorMessage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userGetterMock := mocks.NewUserGetter(t)
			userGetterMock.On("AuthenticateManager", mock.AnythingOfType("*models.User")).
				Return(tc.ManagerID, tc.AuthenticateManagerError).
				Maybe()

			userGetterMock.On("SetSessionID", tc.ManagerID, mock.AnythingOfType("string")).
				Return(tc.SetSessionIDError).
				Maybe()

			reqBody := bytes.NewReader(tc.RequestBody)
			req, err := http.NewRequest(http.MethodPost, loginURL, reqBody)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: test.EnvLocal}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			user.LoginHandler(log, userGetterMock, cfg).ServeHTTP(rr, req)
			require.Equal(t, tc.ResponseCode, rr.Code)

			if tc.ResponseCode == http.StatusAccepted {
				require.True(t, findSessionCookie(rr))
			}

			respBody := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, respBody)

				return
			} else {
				assert.Equal(t, tc.ResponseBody, respBody)

				return
			}
		})
	}
}

// findSessionCookie ищет в полученных куках session_id и проверяет, что он как-то заполнен
func findSessionCookie(rr *httptest.ResponseRecorder) bool {
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "session_id" && cookie.Value != "" {
			return true
		}
	}

	return false
}

func TestCreateManagerHandler(t *testing.T) {
	cases := []struct {
		Name                  string
		CreateNewManagerError error
		RequestBody           []byte
		Environment           string
		ResponseCode          int
		JSON                  bool
		ResponseBody          string
	}{
		{
			Name:                  "Success",
			CreateNewManagerError: nil,
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			ResponseCode: http.StatusCreated,
			JSON:         false,
			ResponseBody: "",
		},
		{
			Name:                  "No login",
			CreateNewManagerError: nil,
			RequestBody: test.MustMarshal(&models.User{
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			ResponseCode: http.StatusBadRequest,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedValidationError, "login", test.Required),
		},
		{
			Name:                  "No password",
			CreateNewManagerError: nil,
			RequestBody: test.MustMarshal(&models.User{
				Login: "aboba",
				Email: "aboba@mail.ru",
			}),
			ResponseCode: http.StatusBadRequest,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedValidationError, "password", test.Required),
		},
		{
			Name:                  "No email",
			CreateNewManagerError: nil,
			RequestBody: test.MustMarshal(&models.User{
				Password: "aboba",
				Login:    "aboba",
			}),
			ResponseCode: http.StatusBadRequest,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedValidationError, "email", test.Required),
		},
		{
			Name: "Login smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "a",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "login", fmt.Sprintf(test.Min, 4)),
		},
		{
			Name: "Login bigger max",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "login", fmt.Sprintf(test.Max, 8)),
		},
		{
			Name: "Password smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "a",
				Email:    "aboba@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "password", fmt.Sprintf(test.Min, 4)),
		},
		{
			Name: "Password bigger max",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Email:    "aboba@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "password", fmt.Sprintf(test.Max, 10)),
		},
		{
			Name: "Email smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "a@m.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "email", fmt.Sprintf(test.Min, 8)),
		},
		{
			Name: "Email bigger max",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody:          fmt.Sprintf(test.ExpectedValidationError, "email", fmt.Sprintf(test.Max, 15)),
		},
		{
			Name: "Validation errors",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Email:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail.ru",
			}),
			CreateNewManagerError: nil,
			ResponseCode:          http.StatusBadRequest,
			JSON:                  true,
			ResponseBody: string(test.MustMarshal(map[string]string{
				"password": fmt.Sprintf(test.Max, 10),
				"email":    fmt.Sprintf(test.Max, 15),
			})),
		},
		{
			Name:                  "Manager already exists",
			CreateNewManagerError: customErr.ErrManagerAlreadyExists,
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			ResponseCode: http.StatusConflict,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedError, "такой менеджер уже существует"),
		},
		{
			Name:                  "Error when creating manager on dev",
			CreateNewManagerError: fmt.Errorf("test error"),
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			Environment:  test.EnvLocal,
			ResponseCode: http.StatusInternalServerError,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedError, "test error"),
		},
		{
			Name:                  "Error when creating manager on prod",
			CreateNewManagerError: fmt.Errorf("test error"),
			RequestBody: test.MustMarshal(&models.User{
				Login:    "aboba",
				Password: "aboba",
				Email:    "aboba@mail.ru",
			}),
			Environment:  test.EnvProd,
			ResponseCode: http.StatusInternalServerError,
			JSON:         false,
			ResponseBody: test.InternalServerErrorMessage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userSetterMock := mocks.NewUserSetter(t)
			userSetterMock.On("CreateNewManager", mock.AnythingOfType("*request.UserCreate")).
				Return(tc.CreateNewManagerError).
				Maybe()

			reqBody := bytes.NewReader(tc.RequestBody)
			req := httptest.NewRequest(http.MethodPost, "/manager", reqBody)

			rr := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: test.EnvLocal}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			user.CreateManagerHandler(log, userSetterMock, cfg).ServeHTTP(rr, req)
			assert.Equal(t, tc.ResponseCode, rr.Code)

			respBody := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, respBody)
			} else {
				assert.Equal(t, tc.ResponseBody, respBody)
			}
		})
	}
}

func TestDeleteManagerHandler(t *testing.T) {
	cases := []struct {
		Name               string
		ManagerID          int
		ManagerIDStr       string
		DeleteManagerError error
		Environment        string
		StatusCode         int
		JSON               bool
		ResponseBody       string
	}{
		{
			Name:               "Success",
			ManagerID:          2,
			DeleteManagerError: nil,
			StatusCode:         http.StatusNoContent,
			JSON:               false,
			ResponseBody:       "",
		},
		{
			Name:               "Not int id on prod",
			ManagerID:          0,
			ManagerIDStr:       "aboba",
			DeleteManagerError: nil,
			Environment:        "prod",
			StatusCode:         http.StatusBadRequest,
			JSON:               true,
			ResponseBody:       fmt.Sprintf(test.ExpectedError, user.InvalidManagerIndex.Error()),
		},
		{
			Name:               "Not int id on dev",
			ManagerID:          0,
			ManagerIDStr:       "aboba",
			DeleteManagerError: nil,
			StatusCode:         http.StatusBadRequest,
			JSON:               true,
			ResponseBody:       fmt.Sprintf(test.ExpectedError, user.InvalidManagerIndex.Error()),
		},
		{
			Name:               "DB error on prod",
			ManagerID:          5,
			DeleteManagerError: fmt.Errorf("aboba"),
			Environment:        "prod",
			StatusCode:         http.StatusInternalServerError,
			JSON:               false,
			ResponseBody:       test.InternalServerErrorMessage,
		},
		{
			Name:               "DB error on dev",
			ManagerID:          5,
			DeleteManagerError: fmt.Errorf("aboba"),
			StatusCode:         http.StatusInternalServerError,
			JSON:               true,
			ResponseBody:       fmt.Sprintf(test.ExpectedError, "aboba"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userSetterMock := mocks.NewUserSetter(t)
			userSetterMock.On("DeleteManager", tc.ManagerID).
				Return(tc.DeleteManagerError).
				Maybe()

			if tc.ManagerIDStr == "" {
				tc.ManagerIDStr = strconv.Itoa(tc.ManagerID)
			}
			r := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, fmt.Sprintf("/manager/%s", tc.ManagerIDStr), nil)

			rec := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			router := chi.NewRouter()
			router.Use(middleware.URLFormat)
			router.Delete("/manager/{id}", user.DeleteManagerHandler(log, userSetterMock, cfg))
			router.ServeHTTP(rec, r)

			assert.Equal(t, tc.StatusCode, rec.Code)

			body := rec.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, body)
				return
			} else {
				assert.Equal(t, tc.ResponseBody, body)
				return
			}
		})
	}
}

func TestGetManagerByIDHandler(t *testing.T) {
	cases := []struct {
		Name             string
		ManagerID        int
		ManagerIDStr     string
		ManagerByID      *models.User
		ManagerByIDError error
		Environment      string
		StatusCode       int
		JSON             bool
		ResponseBody     string
	}{
		{
			Name:      "Success",
			ManagerID: 1,
			ManagerByID: &models.User{
				ID:    1,
				Login: "aboba",
				Email: "aboba@cheer.com",
			},
			StatusCode: http.StatusOK,
			JSON:       true,
			ResponseBody: test.MustMarshalResponse(resp.ManagerResponse{
				ID:    1,
				Login: "aboba",
				Email: "aboba@cheer.com",
				URL:   "/manager/1",
			}),
		},
		{
			Name:         "No such manager",
			ManagerID:    5,
			StatusCode:   http.StatusNotFound,
			ResponseBody: test.NotFound,
		},
		{
			Name:         "Wrong id",
			ManagerIDStr: "aboba",
			Environment:  "prod",
			StatusCode:   http.StatusBadRequest,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedError, user.InvalidManagerIndex.Error()),
		},
		{
			Name:             "DB error on prod",
			ManagerID:        2,
			ManagerByIDError: fmt.Errorf("aboba"),
			Environment:      "prod",
			StatusCode:       http.StatusInternalServerError,
			JSON:             false,
			ResponseBody:     test.InternalServerErrorMessage,
		},
		{
			Name:             "DB error on dev",
			ManagerID:        2,
			ManagerByIDError: fmt.Errorf("aboba"),
			StatusCode:       http.StatusInternalServerError,
			JSON:             true,
			ResponseBody:     fmt.Sprintf(test.ExpectedError, "aboba"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userGetterMock := mocks.NewUserGetter(t)
			userGetterMock.On("GetManagerByID", tc.ManagerID).
				Return(tc.ManagerByID, tc.ManagerByIDError).
				Maybe()

			if tc.ManagerIDStr == "" {
				tc.ManagerIDStr = strconv.Itoa(tc.ManagerID)
			}
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("/manager/%s", tc.ManagerIDStr), nil)

			rec := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()

			cfg := &config.Config{}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			router := chi.NewRouter()
			router.Use(middleware.URLFormat)
			router.Get("/manager/{id}", user.GetManagerByIDHandler(log, userGetterMock, cfg))

			router.ServeHTTP(rec, r)
			assert.Equal(t, tc.StatusCode, rec.Code)

			body := rec.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, body)
				return
			} else {
				assert.Equal(t, tc.ResponseBody, body)
				return
			}
		})
	}
}

func TestGetManagersHandler(t *testing.T) {
	cases := []struct {
		Name             string
		Managers         []*models.User
		GetManagersError error
		Environment      string
		StatusCode       int
		JSON             bool
		ResponseBody     string
	}{
		{
			Name: "Success",
			Managers: []*models.User{
				{
					ID:    1,
					Login: "aboba",
					Email: "aboba@mail.ru",
				},
				{
					ID:    2,
					Login: "aboba2",
					Email: "aboba2@mail.ru",
				},
			},
			GetManagersError: nil,
			StatusCode:       http.StatusOK,
			JSON:             true,
			ResponseBody: test.MustMarshalResponse([]resp.ManagerResponse{
				{
					ID:    1,
					Login: "aboba",
					Email: "aboba@mail.ru",
					URL:   "/manager/1",
				},
				{
					ID:    2,
					Login: "aboba2",
					Email: "aboba2@mail.ru",
					URL:   "/manager/2",
				},
			}),
		},
		{
			Name: "Success with one manager",
			Managers: []*models.User{
				{
					ID:    1,
					Login: "aboba",
					Email: "aboba@mail.ru",
				},
			},
			GetManagersError: nil,
			StatusCode:       http.StatusOK,
			JSON:             true,
			ResponseBody: test.MustMarshalResponse([]resp.ManagerResponse{
				{
					ID:    1,
					Login: "aboba",
					Email: "aboba@mail.ru",
					URL:   "/manager/1",
				},
			}),
		},
		{
			Name:             "No any one",
			Managers:         nil,
			GetManagersError: nil,
			StatusCode:       http.StatusOK,
			JSON:             true,
			ResponseBody:     "[]",
		},
		{
			Name:             "DB error on prod",
			Managers:         nil,
			GetManagersError: fmt.Errorf("aboba"),
			Environment:      "prod",
			StatusCode:       http.StatusInternalServerError,
			JSON:             false,
			ResponseBody:     test.InternalServerErrorMessage,
		},
		{
			Name:             "DB error on dev",
			Managers:         nil,
			GetManagersError: fmt.Errorf("aboba"),
			StatusCode:       http.StatusInternalServerError,
			JSON:             true,
			ResponseBody:     fmt.Sprintf(test.ExpectedError, "aboba"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userGetterMock := mocks.NewUserGetter(t)
			userGetterMock.On("GetManagers").
				Return(tc.Managers, tc.GetManagersError).
				Once()

			r := httptest.NewRequest(http.MethodGet, "/manager", nil)

			rec := httptest.NewRecorder()

			cfg := &config.Config{}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			log := slogdiscard.NewDiscardLogger()

			user.GetManagersHandler(log, userGetterMock, cfg).ServeHTTP(rec, r)
			assert.Equal(t, tc.StatusCode, rec.Code)

			body := rec.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, body)
				return
			} else {
				assert.Equal(t, tc.ResponseBody, body)
				return
			}
		})
	}
}

func TestUpdateManagerHandler(t *testing.T) {

	cases := []struct {
		Name               string
		ManagerID          int
		ManagerIDStr       string
		ManagerUpdated     *models.User
		RequestBody        []byte
		Environment        string
		StatusCode         int
		JSON               bool
		ResponseBody       string
		UpdateManagerError error
	}{
		{
			Name:      "Success",
			ManagerID: 5,
			ManagerUpdated: &models.User{
				ID:       5,
				Login:    "aboba2",
				Password: "aboba2",
				Email:    "aboba2@ab.com",
			},
			RequestBody: test.MustMarshal(map[string]string{
				"login":    "aboba2",
				"email":    "aboba2@ab.com",
				"password": "aboba2",
			}),
			StatusCode: http.StatusOK,
			JSON:       true,
			ResponseBody: test.MustMarshalResponse(&resp.ManagerResponse{
				ID:    5,
				Login: "aboba2",
				Email: "aboba2@ab.com",
				URL:   "/manager/5",
			}),
		},
		{
			Name:      "Wrong validation",
			ManagerID: 5,
			ManagerUpdated: &models.User{
				ID:       5,
				Login:    "aboba2",
				Password: "aboba2",
				Email:    "aboba2@ab.com",
			},
			RequestBody: test.MustMarshal(map[string]string{
				"login": "aaaaaaaaaaaaaaaaaaaaaa",
			}),
			StatusCode:   http.StatusBadRequest,
			JSON:         true,
			ResponseBody: fmt.Sprintf(test.ExpectedValidationError, "login", fmt.Sprintf(test.Max, 8)),
		},
		{
			Name:      "Update error on dev",
			ManagerID: 5,
			ManagerUpdated: &models.User{
				ID:       5,
				Login:    "aboba2",
				Password: "aboba2",
				Email:    "aboba2@ab.com",
			},
			RequestBody: test.MustMarshal(map[string]string{
				"login":    "aboba2",
				"email":    "aboba2@ab.com",
				"password": "aboba2",
			}),
			UpdateManagerError: fmt.Errorf("aboba"),
			StatusCode:         http.StatusInternalServerError,
			JSON:               true,
			ResponseBody:       fmt.Sprintf(test.ExpectedError, "aboba"),
		},
		{
			Name:      "Update error on prod",
			ManagerID: 5,
			ManagerUpdated: &models.User{
				ID:       5,
				Login:    "aboba2",
				Password: "aboba2",
				Email:    "aboba2@ab.com",
			},
			RequestBody: test.MustMarshal(map[string]string{
				"login":    "aboba2",
				"email":    "aboba2@ab.com",
				"password": "aboba2",
			}),
			UpdateManagerError: fmt.Errorf("aboba"),
			Environment:        "prod",
			StatusCode:         http.StatusInternalServerError,
			JSON:               false,
			ResponseBody:       test.InternalServerErrorMessage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			userSetterMock := mocks.NewUserSetter(t)
			userSetterMock.On("GetManagerByID", tc.ManagerID).
				Return(tc.ManagerUpdated, nil).
				Maybe()

			userSetterMock.On("UpdateManager", mock.AnythingOfType("*user.UserPatch")).
				Return(tc.UpdateManagerError).
				Maybe()

			reqBody := bytes.NewReader(tc.RequestBody)

			if tc.ManagerIDStr == "" {
				tc.ManagerIDStr = strconv.Itoa(tc.ManagerID)
			}
			r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/manager/%s", tc.ManagerIDStr), reqBody)

			rec := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()

			cfg := &config.Config{}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			router := chi.NewRouter()
			router.Use(middleware.URLFormat)
			router.Patch("/manager/{id}", user.UpdateManagerHandler(log, userSetterMock, cfg))

			router.ServeHTTP(rec, r)
			assert.Equal(t, tc.StatusCode, rec.Code)

			respBody := rec.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, respBody)
			} else {
				assert.Equal(t, tc.ResponseBody, respBody)
			}
		})
	}
}
