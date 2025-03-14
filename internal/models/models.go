package models

import "strings"

// Parking - данные о парковке.
type Parking struct {
	ID      int             `json:"id,omitempty"`
	Name    string          `json:"name" validate:"required,min=3,max=10"`
	Address string          `json:"address" validate:"required,min=10,max=30"`
	Width   int             `json:"width" validate:"required,gte=4,lte=6"`
	Height  int             `json:"height" validate:"required,gte=4,lte=6"`
	Cells   [][]ParkingCell `json:"cells,omitempty"`
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

type ParkingCellStruct struct {
	X, Y     int
	CellType ParkingCell
}

type Manager struct {
	ID       int
	Login    string
	Password string
	Email    string
}
