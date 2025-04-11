package response

import (
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

type Response struct {
	Error string `json:"error,omitempty"`
}

type RespErrorList struct {
	ErrorList []string `json:"error"`
}

// ParkingResponse - формат информации для response об одной парковке.
type ParkingResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	DayTariff   int    `json:"day_tariff"`
	NightTariff int    `json:"night_tariff"`
	URL         string `json:"url"`
}

// UnknownError - ответ, возвращаемый без конкретного поля ошибки.
func UnknownError(errMessage string) map[string]interface{} {
	return map[string]interface{}{
		"error": errMessage,
	}
}

func Error(field string, err error) map[string]interface{} {
	return map[string]interface{}{
		field: err.Error(),
	}
}

// ListError создает массив ошибок.
func ListError(field string, errors []error) map[string]interface{} {
	var errMessageList []string

	for _, err := range errors {
		errMessageList = append(errMessageList, err.Error())
	}

	return map[string]interface{}{
		field: errMessageList,
	}
}

// NewParkingResponse создает ответ ParkingResponse для рендера.
func NewParkingResponse(p *models.Parking) *ParkingResponse {
	return &ParkingResponse{
		ID:          p.ID,
		Name:        p.Name,
		Address:     p.Address,
		DayTariff:   p.DayTariff,
		NightTariff: p.NightTariff,
		URL:         fmt.Sprintf("/parking/%d", p.ID),
	}
}

// Render нужна для имплементации интерфейса Renderer.
func (*ParkingResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// NewParkingListRender подготавливает информацию о парковках к выводу.
func NewParkingListRender(parkings []*models.Parking) []render.Renderer {
	var list []render.Renderer

	for _, parking := range parkings {
		pResponse := NewParkingResponse(parking)
		list = append(list, pResponse)
	}

	return list
}

// validationErrorMessages нужна для получения стандартного сообщения ошибки
// для тегов ошибок типа validator.ValidationErrors.
var validationErrorMessages = map[string]string{
	"required":           "Не указано поле",
	"min":                "Минимальная длина поля %s",
	"max":                "Максимальная длина поля %s",
	"lte":                "Значение не может быть больше %s",
	"gte":                "Значение не может быть меньше %s",
	"email":              "Введенное значение не email",
	"required_with_type": "Необходимо вместе с %s",
}

func ValidationError(validateErr validator.ValidationErrors) map[string]string {
	fieldErrors := make(map[string]string)

	for _, err := range validateErr {
		field := err.Field()
		tag := err.Tag()
		param := err.Param()

		var errMessage string
		if param != "" {
			errMessage = fmt.Sprintf(validationErrorMessages[tag], param)
		} else {
			errMessage = validationErrorMessages[tag]
		}

		if errMessage != "" {
			fieldErrors[field] = errMessage
		}
	}

	return fieldErrors
}

func RecursiveValidationError(validateErr validator.ValidationErrors) map[string]interface{} {
	fieldErrors := make(map[string]interface{})

	for _, err := range validateErr {
		namespace := err.Namespace()
		path := strings.Split(namespace, ".")

		tag := err.Tag()
		param := err.Param()

		var errMessage string
		if param != "" {
			errMessage = fmt.Sprintf(validationErrorMessages[tag], param)
		} else {
			errMessage = validationErrorMessages[tag]
		}

		if errMessage != "" {
			insertFieldError(fieldErrors, path, errMessage)
		}
	}

	return fieldErrors
}

func insertFieldError(m map[string]interface{}, path []string, message string) {
	if len(path) == 1 {
		m[path[0]] = message
		return
	}

	key := path[0]

	if _, ok := m[key]; !ok {
		m[key] = make(map[string]interface{})
	}

	subMap, _ := m[key].(map[string]interface{})
	insertFieldError(subMap, path[1:], message)
}

// ErrorHandler обрабатывает серверную ошибку (не клиентскую).
// Если приложение находится не в проде, выведет ошибку пользователю.
// Иначе, выведет стандартное сообщение "Internal Server UnknownError".
func ErrorHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, err error) {
	if cfg.Environment != "prod" {
		renderError(w, r, err)
	} else {
		internalError(w)
	}
}

// internalError возвращает ошибку сервера без дополнительной информации для пользователя.
func internalError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// renderError предоставляет текст ошибки пользователя. Используется в версии для разработки.
func renderError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, UnknownError(err.Error()))
}
