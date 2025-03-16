package user_test

import (
	"bytes"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/user"
	"github.com/PIRSON21/parking/internal/http-server/handler/user/mocks"
	customErr "github.com/PIRSON21/parking/internal/lib/errors"
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

			assert.Fail(t, "не все проверки прописаны")
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

const (
	createManagerURL = "/create_manager"
)

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
		//{
		//	Name:                  "No email",
		//	CreateNewManagerError: nil,
		//	RequestBody: test.MustMarshal(&models.User{
		//		Password: "aboba",
		//		Login:    "aboba",
		//	}),
		//	ResponseCode: http.StatusBadRequest,
		//	JSON:         true,
		//	ResponseBody: fmt.Sprintf(test.ExpectedValidationError, "email", test.Required),
		//},
		// TODO: не работает сейчас
		{
			Name: "Login smaller min",
			RequestBody: test.MustMarshal(&models.User{
				Login:    "a",
				Password: "aboba",
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
			ResponseCode: http.StatusInternalServerError,
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
			userSetterMock.On("CreateNewManager", mock.AnythingOfType("*models.User")).
				Return(tc.CreateNewManagerError).
				Maybe()

			reqBody := bytes.NewReader(tc.RequestBody)
			req := httptest.NewRequest(http.MethodPost, createManagerURL, reqBody)

			rr := httptest.NewRecorder()

			log := slogdiscard.NewDiscardLogger()
			cfg := &config.Config{Environment: test.EnvLocal}
			if tc.Environment != "" {
				cfg.Environment = tc.Environment
			}

			user.CreateManagerHandler(log, userSetterMock, cfg).ServeHTTP(rr, req)
			require.Equal(t, tc.ResponseCode, rr.Code)

			respBody := rr.Body.String()

			if tc.JSON {
				assert.JSONEq(t, tc.ResponseBody, respBody)

				return
			} else {
				assert.Equal(t, tc.ResponseBody, respBody)

				return
			}

			assert.Fail(t, "не все тесты прописаны")
		})
	}
}
