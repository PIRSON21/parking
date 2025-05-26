package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PIRSON21/parking/internal/models"
	"github.com/PIRSON21/parking/internal/simulation"
	"github.com/go-playground/validator/v10"
	"golang.org/x/xerrors"
)

// CreateNewValidator создает объект типа *validator.Validate, в котором название поля берется из json тега.
func CreateNewValidator() *validator.Validate {
	valid := validator.New()

	valid.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return valid
}

func ArrivalConfigStructLevelValidation(sl validator.StructLevel) {
	ac := sl.Current().Interface().(simulation.ArrivalConfig)

	switch ac.Type {
	case "normal":
		if ac.Mean == 0 {
			sl.ReportError(ac.Mean, "mean", "mean", "required_with_type", "normal")
		}
		if ac.StdDev == 0 {
			sl.ReportError(ac.StdDev, "std_dev", "std_dev", "required_with_type", "normal")
		}
	case "exponential":
		if ac.Lambda == 0 {
			sl.ReportError(ac.Lambda, "lambda", "lambda", "required_with_type", "exponential")
		}
	case "uniform":
		if ac.MinDelay == 0 || ac.MaxDelay == 0 {
			sl.ReportError(ac.MinDelay, "min_delay", "min_delay", "required_with_type", "uniform")
			sl.ReportError(ac.MaxDelay, "max_delay", "max_delay", "required_with_type", "uniform")
		}
		if ac.MaxDelay < ac.MinDelay {
			sl.ReportError(ac.MinDelay, "min_delay", "MinDelay", "lte", fmt.Sprintf("%0.2f", ac.MaxDelay))
		}
	case "discrete":
		if ac.DiscreteTime == 0 {
			sl.ReportError(ac.DiscreteTime, "DiscreteTime", "discrete_time", "required_with_type", "discrete")
		}
	}
}

func ParkingTimeConfigStructLevelValidation(sl validator.StructLevel) {
	tc := sl.Current().Interface().(simulation.ParkingTimeConfig)

	switch tc.Type {
	case "normal":
		if tc.Mean == 0 {
			sl.ReportError(tc.Mean, "mean", "mean", "required_with_type", "normal")
		}
		if tc.StdDev == 0 {
			sl.ReportError(tc.StdDev, "std_dev", "std_dev", "required_with_type", "normal")
		}
	case "exponential":
		if tc.Lambda == 0 {
			sl.ReportError(tc.Lambda, "lambda", "lambda", "required_with_type", "exponential")
		}
	case "uniform":
		if tc.MinDuration == 0 || tc.MaxDuration == 0 {
			sl.ReportError(tc.MinDuration, "min_delay", "min_delay", "required_with_type", "uniform")
			sl.ReportError(tc.MaxDuration, "max_delay", "max_delay", "required_with_type", "uniform")
		}
		if tc.MaxDuration < tc.MinDuration {
			sl.ReportError(tc.MinDuration, "min_delay", "MinDelay", "lte", fmt.Sprintf("%0.2f", tc.MaxDuration))
		}
	case "discrete":
		if tc.DiscreteTime == 0 {
			sl.ReportError(tc.DiscreteTime, "DiscreteTime", "discrete_time", "required_with_type", "discrete")
		}
	}
}

// ValidateParkingCells проверяет клетки парковки на соответствие требованиям.
// Возвращает список всех найденных ошибок
func ValidateParkingCells(parking *models.Parking) []error {
	var errors []error
	var countEnterance, countExit int

	height := len(parking.Cells)
	if height != parking.Height {
		errors = append(errors, xerrors.Errorf("длина парковки не соответствует длине топологии: %d", parking.Height))
		return errors
	}

	for i, width := range parking.Cells {
		if len(width) != parking.Width {
			errors = append(errors, xerrors.Errorf("ширина строки %d не соответствует ширине топологии: %d", i, parking.Width))
		}

		for j, cell := range width {
			if !cell.IsParkingCell() {
				errors = append(errors, xerrors.Errorf("клетка (%d,%d) недействительна: '%s'", j, i, cell))
			} else if cell.IsEntrance() {
				errors = append(errors, validateEnterance(&countEnterance, i, height, j)...)
			} else if cell.IsExit() {
				errors = append(errors, validateExit(&countExit, height, i, j)...)
			}
		}
	}

	if len(errors) != 0 {
		return errors
	}

	return nil
}

func validateExit(countExit *int, height int, i int, j int) []error {
	var errors []error
	if *countExit >= 1 {
		errors = append(errors, xerrors.Errorf("в топологии парковки не может быть более одной точки выхода"))
	}
	*countExit++
	if i != height-1 {
		errors = append(errors, xerrors.Errorf("точка выхода должна быть в нижней строке парковки, а не на (%d,%d)", i, j))
	}
	return errors
}

func validateEnterance(countEnterance *int, i int, height int, j int) []error {
	var errors []error
	if *countEnterance >= 1 {
		errors = append(errors, xerrors.Errorf("в топологии парковки не может быть более одной точки входа"))
	}
	*countEnterance++
	if i != height-1 {
		errors = append(errors, xerrors.Errorf("точка входа должна быть в нижней строке парковки, а не на (%d,%d)", i, j))
	}
	return errors
}
