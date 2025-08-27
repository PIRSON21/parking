package simulation

import (
	"encoding/json"
	"log"

	"github.com/PIRSON21/parking/internal/models"
	"github.com/google/uuid"
	customTimer "github.com/ivahaev/timer"
)

// CarEvent - тело события машины (прибытие, парковка, отъезд)
type CarEvent struct {
	Event     string   `json:"event"`             // "arrive", "drove-away", "leave"
	CarID     string   `json:"car_id"`            // id машины
	TimeStamp int64    `json:"timestamp"`         // время события
	ParkID    *int     `json:"park_id,omitempty"` // id парковочного места
	ParkX     *int     `json:"park_x,omitempty"`  // х координата парковочного места
	ParkY     *int     `json:"park_y,omitempty"`  // y координата парковочного места
	Price     *float64 `json:"price,omitempty"`   // стоимость парковки
}

const (
	eventArrive = "arrive" // eventArrive - машина появляется на дороге
)

// generateCarID создает уникальный id машины.
func generateCarID() string {
	return uuid.New().String()
}

// scheduleCar создает новое событие появления автомобиля
func (ss *Session) scheduleCar() {
	delay := ss.generateArrivalDelay()
	t := customTimer.NewTimer(delay)
	t.Start()

	for {
		select {
		case <-t.C:
			ss.mu.Lock()
			if ss.isRunning() {
				ss.mu.Unlock()
				go ss.sendCarEvent()
				delay = ss.generateArrivalDelay()
				t = customTimer.NewTimer(delay)
				t.Start()
			} else {
				ss.mu.Unlock()
				return
			}
		case <-ss.ctx.Done():
			ss.log.Debug("session stopped, stop scheduling cars")
			return
		}
	}
}

// sendCarEvent создает событие о появлении автомобиля.
func (ss *Session) sendCarEvent() {
	err := ss.sem.Acquire(ss.ctx, 1)
	defer ss.sem.Release(1)
	if err != nil {
		return
	}
	carID := generateCarID()

	car := &models.SimulatedCar{
		CarID: carID,
		State: eventArrive,
	}

	ss.mu.Lock()
	ss.car[carID] = car
	ss.mu.Unlock()

	event := CarEvent{
		Event:     eventArrive,
		CarID:     carID,
		TimeStamp: ss.timer.elapsedTime.Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Println("error while marshaling: ", err)
		return
	}

	ss.log.Debug("car arrived", "car_id", carID, "time", ss.timer.elapsedTime)
	ss.client.Send(data)
}

// tryToPark определяет, заедет машина на парковку или нет.
func (ss *Session) tryToPark(carID string) {
	err := ss.sem.Acquire(ss.ctx, 1)
	defer ss.sem.Release(1)
	if err != nil {
		return
	}

	ss.mu.Lock()
	if !ss.isRunning() {
		return
	}

	ss.log.Debug("trying to park car", "car_id", carID)

	car, ok := ss.car[carID]
	if !ok {
		ss.mu.Unlock()
		return
	}
	ss.mu.Unlock()

	if canEnter := ss.evaluateEntrance(); !canEnter {
		ss.droveAwayCar(carID)
		return
	}

	spot, ok := ss.parking.OccupySpot()
	if !ok {
		ss.droveAwayCar(carID)
		return
	}

	ss.mu.Lock()
	car.State = "park"
	car.Spot = spot
	ss.mu.Unlock()
	ss.sendParkEvent(carID)
}

func (ss *Session) sendParkEvent(carID string) {
	ss.mu.Lock()

	car, ok := ss.car[carID]
	if !ok || car == nil {
		ss.log.Debug("car not found in session", "car_id", carID)
		ss.mu.Unlock()
		return
	}

	car.EnterTime = ss.timer.elapsedTime
	ss.mu.Unlock()

	event := CarEvent{
		Event:     "park",
		CarID:     carID,
		ParkX:     &car.Spot.X,
		ParkY:     &car.Spot.Y,
		TimeStamp: ss.timer.elapsedTime.Unix(),
	}

	ss.log.Debug("car parked", "car_id", carID, "time", ss.timer.elapsedTime, "spot", car.Spot)

	ss.eventChan <- event

	ss.scheduleLeave(carID, car.Spot)
}

// droveAwayCar создает событие, когда автомобиль не заезжает на парковку.
func (ss *Session) droveAwayCar(carID string) {
	err := ss.sem.Acquire(ss.ctx, 1)
	if err != nil {
		return
	}
	defer ss.sem.Release(1)

	event := CarEvent{
		Event:     "drove-away",
		CarID:     carID,
		TimeStamp: ss.timer.elapsedTime.Unix(),
	}

	delete(ss.car, carID)

	ss.log.Debug("car drove away", "car_id", carID, "time", ss.timer.elapsedTime)

	ss.eventChan <- event
}

// scheduleLeave планирует выезд автомобиля.
func (ss *Session) scheduleLeave(carID string, spot *models.ParkingPoint) {
	if !ss.isRunning() {
		return
	}

	delay := ss.generateLeaveDelay()

	ss.mu.Lock()
	car := ss.car[carID]

	time := customTimer.AfterFunc(delay, func() {
		ss.mu.Lock()

		if !ss.isRunning() || ss.ctx.Err() != nil {
			ss.log.Debug("session stopped or context cancelled, stop leaving car", "car_id", carID)
			ss.mu.Unlock()
			return
		}

		car.Price = ss.calculateParkingCost(ss.timer.elapsedTime, car.EnterTime)

		ss.parking.ReleaseSpot(spot)

		ss.sendLeaveParkEvent(carID)
		ss.mu.Unlock()
	})

	car.Timer = time
	ss.mu.Unlock()
	time.Start()
}

// sendLeaveParkEvent отправляет событие о выезде автомобиля с парковки.
func (ss *Session) sendLeaveParkEvent(carID string) {
	car := ss.car[carID]

	event := CarEvent{
		Event:     "leave",
		CarID:     carID,
		ParkX:     &car.Spot.X,
		ParkY:     &car.Spot.Y,
		TimeStamp: ss.timer.elapsedTime.Unix(),
		Price:     &car.Price,
	}

	delete(ss.car, carID)

	ss.log.Debug("car left parking", "car_id", carID, "time", ss.timer.elapsedTime, "price", car.Price, "spot", car.Spot)

	ss.eventChan <- event
}

func (ss *Session) eventLoop() {
	for {
		select {

		case event, ok := <-ss.eventChan:
			if !ok {
				ss.log.Debug("event channel closed, stopping event loop")
				return
			}

			data, err := json.Marshal(&event)
			if err != nil {
				log.Println("error while marshaling event: ", err)
				continue
			}

			ss.log.Debug("sending event", "event", event.Event, "car_id", event.CarID, "time", event.TimeStamp)

			ss.client.Send(data)
		case <-ss.ctx.Done():
			ss.log.Debug("session stopped, stopping event loop")
			return
		}
	}
}
