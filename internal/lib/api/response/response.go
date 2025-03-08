package response

import (
	"github.com/PIRSON21/parking/internal/models"
	"github.com/go-chi/render"
	"net/http"
)

type Response struct {
	Error string `json:"error,omitempty"`
}

// ParkingResponse - формат информации для response об одной парковке.
type ParkingResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Height  int    `json:"height"`
	Width   int    `json:"width"`
}

func Error(errMessage string) Response {
	return Response{
		Error: errMessage,
	}
}

// NewParkingResponse создает ответ ParkingResponse для рендера.
func NewParkingResponse(p *models.Parking) *ParkingResponse {
	return &ParkingResponse{
		ID:      p.ID,
		Name:    p.Name,
		Address: p.Address,
		Height:  p.Height,
		Width:   p.Width,
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
