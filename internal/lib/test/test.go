package test

import "encoding/json"

func MustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func MustMarshalResponse(v interface{}) string {
	var res []byte
	res, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(res)
}

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"

	Required = "Не указано поле"
	Min      = "Минимальная длина поля %d"
	Max      = "Максимальная длина поля %d"
	Lte      = "Значение не может быть больше %d"
	Gte      = "Значение не может быть меньше %d"

	ExpectedError            = `{"error":%q}`
	ExpectedErrors           = `{"error":[%s]}`
	ExpectedValidationError  = `{%q:%q}`
	ExpectedValidationErrors = `{%q:[%s]}`

	InternalServerErrorMessage = "Internal Server Error\n"

	NotFound = "404 page not found\n"
)
