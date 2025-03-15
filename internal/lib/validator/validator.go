package validator

import (
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
