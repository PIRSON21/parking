package middleware_test

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	"github.com/PIRSON21/parking/internal/lib/api/auth/middleware/mocks"
	custErr "github.com/PIRSON21/parking/internal/lib/errors"
	"github.com/PIRSON21/parking/internal/lib/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	expectedUnauthorizedError   = "Unauthorized\n"
	expectedExpiredSessionError = "Session Expired\n"
)

func TestAuthMiddleware(t *testing.T) {
	cases := []struct {
		Name           string
		SessionID      string
		UserID         int
		GetUserIDError error
		Cookie         *http.Cookie
		ResponseCode   int
		ResponseBody   string
	}{
		{
			Name:           "Success admin",
			SessionID:      "aboba",
			UserID:         0,
			GetUserIDError: nil,
			Cookie: &http.Cookie{
				Name:  "session_id",
				Value: "aboba",
			},
			ResponseCode: http.StatusOK,
			ResponseBody: "",
		},
		{
			Name:           "Success manager",
			SessionID:      "aboba",
			UserID:         5,
			GetUserIDError: nil,
			Cookie: &http.Cookie{
				Name:  "session_id",
				Value: "aboba",
			},
			ResponseCode: http.StatusOK,
			ResponseBody: "",
		},
		{
			Name:           "No cookie",
			SessionID:      "",
			UserID:         0,
			GetUserIDError: nil,
			Cookie:         nil,
			ResponseCode:   http.StatusUnauthorized,
			ResponseBody:   expectedUnauthorizedError,
		},
		{
			Name:           "No session cookie",
			SessionID:      "",
			UserID:         0,
			GetUserIDError: nil,
			Cookie: &http.Cookie{
				Name:  "wrong_cookie",
				Value: "aboba",
			},
			ResponseCode: http.StatusUnauthorized,
			ResponseBody: expectedUnauthorizedError,
		},
		{
			Name:           "No session on DB",
			SessionID:      "aboba",
			UserID:         0,
			GetUserIDError: custErr.ErrUnauthorized,
			Cookie: &http.Cookie{
				Name:  "session_id",
				Value: "aboba",
			},
			ResponseCode: http.StatusUnauthorized,
			ResponseBody: expectedUnauthorizedError,
		},
		{
			Name:           "Expired session on DB",
			SessionID:      "aboba",
			UserID:         0,
			GetUserIDError: custErr.ErrSessionExpired,
			Cookie: &http.Cookie{
				Name:  "session_id",
				Value: "aboba",
			},
			ResponseCode: http.StatusForbidden,
			ResponseBody: expectedExpiredSessionError,
		},
		{
			Name:           "Internal error",
			SessionID:      "aboba",
			UserID:         0,
			GetUserIDError: fmt.Errorf("test middleware error"),
			Cookie: &http.Cookie{
				Name:  "session_id",
				Value: "aboba",
			},
			ResponseCode: http.StatusInternalServerError,
			ResponseBody: test.InternalServerErrorMessage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(middleware.UserIDKey)
				require.NotEqual(t, userID, nil)

				valInt, ok := userID.(int)
				require.True(t, ok)

				assert.Equal(t, tc.UserID, valInt)
			})

			authGetterMock := mocks.NewAuthGetter(t)
			authGetterMock.On("GetUserID", tc.SessionID).
				Return(tc.UserID, tc.GetUserIDError).
				Maybe()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.Cookie != nil {
				req.AddCookie(tc.Cookie)
			}

			rr := httptest.NewRecorder()

			middleware.AuthMiddleware(authGetterMock)(nextHandler).ServeHTTP(rr, req)

			require.Equal(t, tc.ResponseCode, rr.Code)

			body := rr.Body.String()

			assert.Equal(t, tc.ResponseBody, body)
		})
	}
}
