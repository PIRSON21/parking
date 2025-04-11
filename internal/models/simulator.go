package models

import (
	"math"
	"sync"
	"time"

	"github.com/ivahaev/timer"
)

// ParkingSpot отражает парковочное место.
type ParkingSpot struct {
	ID     int
	IsFree bool
	X, Y   int
}

// ParkingLot отражает топологию парковки.
type ParkingLot struct {
	mu     sync.Mutex
	spots  []*ParkingSpot
	EntryX int
	EntryY int
}

// NewParkingLot создает модель парковки для сессии.
func NewParkingLot(parking *Parking) *ParkingLot {
	count := 0

	var spots []*ParkingSpot
	var entryX, entryY int

	for width := 0; width < len(parking.Cells); width++ {
		for height := 0; height < len(parking.Cells[width]); height++ {
			if parking.Cells[width][height].IsParking() {
				parkingSpot := &ParkingSpot{
					ID:     count,
					IsFree: true,
					X:      width,
					Y:      height,
				}
				spots = append(spots, parkingSpot)
				count++
			} else if parking.Cells[width][height].IsEntrance() {
				entryX = width
				entryY = height
			}
		}
	}

	return &ParkingLot{
		spots:  spots,
		EntryX: entryX,
		EntryY: entryY,
	}
}

// HasFreeSpot проверит парковку на свободные места.
func (p *ParkingLot) HasFreeSpot() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, spot := range p.spots {
		if spot.IsFree {
			return true
		}
	}
	return false
}

// OccupySpot занимает парковочное место (если найдёт). Вернёт nil, false если места нет.
func (p *ParkingLot) OccupySpot() (*ParkingSpot, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var nearestSpot *ParkingSpot
	minDistance := math.MaxFloat64

	nearestSpot = p.findNearestSpot(minDistance, nearestSpot)

	if nearestSpot != nil {
		nearestSpot.IsFree = false
		return nearestSpot, true
	}

	return nil, false
}

// findNearestSpot находит ближайшее парковочное место по расстоянию по Манхэттену.
func (p *ParkingLot) findNearestSpot(minDistance float64, nearestSpot *ParkingSpot) *ParkingSpot {
	for _, spot := range p.spots {
		if spot.IsFree {
			distance := math.Abs(float64(p.EntryX-spot.X)) + math.Abs(float64(p.EntryY-spot.Y))

			if distance < minDistance {
				minDistance = distance
				nearestSpot = spot
			}
		}
	}
	return nearestSpot
}

// ReleaseSpot освобождает занятое парковочное место.
func (p *ParkingLot) ReleaseSpot(spot *ParkingSpot) {
	p.mu.Lock()
	defer p.mu.Unlock()

	spot.IsFree = true
}

// SimulatedCar описывает машину в симуляции.
type SimulatedCar struct {
	CarID     string
	State     string
	Timer     *timer.Timer
	Spot      *ParkingSpot
	EnterTime time.Time
	Price     int
}
