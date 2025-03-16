package errors

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

var ErrSessionExpired = errors.New("session expired")

var ErrManagerAlreadyExists = errors.New("такой менеджер уже существует")
