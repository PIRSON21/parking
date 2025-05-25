package models

import (
	"strings"
)

// Parking - данные о парковке.
type Parking struct {
	ID          int             `json:"id,omitempty"`
	Name        string          `json:"name" validate:"required,min=3,max=10"`
	Address     string          `json:"address" validate:"required,min=10,max=30"`
	Width       int             `json:"width" validate:"required,gte=4,lte=6"`
	Height      int             `json:"height" validate:"required,gte=4,lte=6"`
	DayTariff   *int            `json:"day_tariff" validate:"required,gte=0,lte=1000"`
	NightTariff *int            `json:"night_tariff" validate:"required,gte=0,lte=1000"`
	Cells       [][]ParkingCell `json:"cells,omitempty"`
	Manager     *Manager        `json:"manager,omitempty"`
}

type Manager struct {
	ID int `json:"id"`
}

// ParkingCell - строка, которая хранит в себе информацию о клетки парковки
type ParkingCell string

const (
	Road       ParkingCell = "."
	Park       ParkingCell = "P"
	Entrance   ParkingCell = "I"
	Exit       ParkingCell = "O"
	Decoration ParkingCell = "D"
)

// validCells - мапа для проверки клетки (так быстрее).
var validCells = map[ParkingCell]struct{}{
	Road:       {},
	Park:       {},
	Entrance:   {},
	Exit:       {},
	Decoration: {},
}

// IsParkingCell проверяет, является ли текущая строка - правильной ParkingCell.
func (c *ParkingCell) IsParkingCell() bool {
	_, exists := validCells[*c]
	return exists
}

// IsRoad проверяет, является ли текущая клетка - дорогой.
func (c *ParkingCell) IsRoad() bool {
	return strings.EqualFold(string(*c), string(Road))
}

func (c *ParkingCell) IsParking() bool {
	return strings.EqualFold(string(*c), string(Park))
}

func (c *ParkingCell) IsEntrance() bool {
	return *c == Entrance
}

func (c *ParkingCell) IsExit() bool {
	return *c == Exit
}

// ParkingCellStruct используется для получения/сохранения данных о клетках парковки в БД
type ParkingCellStruct struct {
	X, Y     int
	CellType ParkingCell
}

// User отражает поля пользователей
type User struct {
	ID       int
	Login    string `json:"login" validate:"required,min=4,max=8"`
	Password string `json:"password" validate:"required,min=4,max=10"`
	Email    string `json:"email,omitempty" validate:"omitempty,email,min=8,max=15"`
}
