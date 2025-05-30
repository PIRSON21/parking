package errors

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

var ErrSessionExpired = errors.New("session expired")

var ErrManagerAlreadyExists = errors.New("такой менеджер уже существует")

var ErrManagerNotFound = errors.New("менеджер не найден")

var ErrParkingNotFound = errors.New("парковка не найдена")

var ErrParkingAccessDenied = errors.New("доступ к парковке запрещен")

var ErrParkingAlreadyExists = errors.New("парковка с таким именем и адресом уже существует")
