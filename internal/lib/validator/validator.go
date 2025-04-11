package validator

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/simulation"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
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
