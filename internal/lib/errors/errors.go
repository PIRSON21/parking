package errors

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

var ErrSessionExpired = errors.New("session expired")
