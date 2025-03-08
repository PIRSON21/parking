package models

// Parking - данные о парковке.
type Parking struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name" validate:"required,min=3,max=10"`
	Address string `json:"address" validate:"required,min=10,max=30"`
	Width   int    `json:"width" validate:"required,gte=4,lte=6"`
	Height  int    `json:"height" validate:"required,gte=4,lte=6"`
}
