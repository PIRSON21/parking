package simulation

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
	"strings"
	"sync"
	"time"

	"github.com/PIRSON21/parking/internal/models"
)

type EventSender interface {
	Send(data []byte)
}

// Session описывает сессию пользователя.
type Session struct {
	mu         sync.Mutex
	state      string // "running", "paused", "stopped"
	ctx        context.Context
	cancel     context.CancelFunc
	client     EventSender
	parking    *models.ParkingLot
	car        map[string]*models.SimulatedCar
	timer      *Timer
	sem        *semaphore.Weighted
	arrivalCfg *ArrivalConfig
	parkingCfg *ParkingTimeConfig
	pauseCh    chan struct{}
}

// Timer имеет данные о времени симуляции.
type Timer struct {
	startTime   time.Time
	elapsedTime time.Time
	ticker      *time.Ticker
}

// ArrivalConfig описывает данные моделирования.
type ArrivalConfig struct {
	Type         string  `json:"type" validate:"oneof=exponential normal uniform discrete"` // "exponential", "normal", "uniform", "discrete"
	Lambda       float64 `json:"lambda,omitempty" validate:"omitempty,lte=1,gte=0.1"`       // Для экспоненциального распределения (интенсивность)
	Mean         float64 `json:"mean,omitempty" validate:"omitempty,lte=15,gte=2"`          // Среднее значение для нормального распределения
	StdDev       float64 `json:"std_dev,omitempty" validate:"omitempty,lte=15,gte=0.1"`     // Стандартное отклонение для нормального распределения
	MinDelay     float64 `json:"min_delay,omitempty" validate:"omitempty,lte=15,gte=2"`     // Минимальная задержка для равномерного распределения
	MaxDelay     float64 `json:"max_delay,omitempty" validate:"omitempty,lte=15,gte=2"`     // Максимальная задержка для равномерного распределения
	DiscreteTime float64 `json:"discrete_time,omitempty" validate:"omitempty"`              // Время появления для дискретного типа
	ParkingProb  float64 `json:"parking_prob" validate:"required,lte=1,gte=0"`              // Вероятность заезда автомобиля на парковку
}

type ParkingTimeConfig struct {
	Type         string  `json:"type" validate:"oneof=exponential normal uniform discrete"` // "exponential", "normal", "uniform", "discrete"
	Lambda       float64 `json:"lambda,omitempty" validate:"omitempty,lte=1,gte=0.1"`       // Для экспоненциального распределения
	Mean         float64 `json:"mean,omitempty" validate:"omitempty,lte=15,gte=2"`          // Среднее время стоянки
	StdDev       float64 `json:"std_dev,omitempty" validate:"omitempty,lte=15,gte=0.1"`     // Стандартное отклонение для нормального распределения
	MinDuration  float64 `json:"min_delay,omitempty" validate:"omitempty,lte=15,gte=2"`     // Минимальная длительность для равномерного распределения
	MaxDuration  float64 `json:"max_delay,omitempty" validate:"omitempty,lte=15,gte=2"`     // Максимальная длительность для равномерного распределения
	DiscreteTime float64 `json:"discrete_time,omitempty" validate:"omitempty"`              // Дискретное значение длительности стоянки
}

const (
	stateRunning = "running"
	statePaused  = "paused"
	stateStopped = "stopped"
)

func NewSession(client EventSender, parking *models.Parking, startTime time.Time, arrivalCfg *ArrivalConfig, parkingCfg *ParkingTimeConfig) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	parkingLot := models.NewParkingLot(parking)
	sem := semaphore.NewWeighted(20)

	if startTime.IsZero() {
		startTime = time.Now()
	}

	return &Session{
		state:   stateStopped,
		ctx:     ctx,
		cancel:  cancel,
		client:  client,
		parking: parkingLot,
		car:     make(map[string]*models.SimulatedCar),
		timer: &Timer{
			startTime:   startTime,
			elapsedTime: startTime,
			ticker:      nil,
		},
		sem:        sem,
		arrivalCfg: arrivalCfg,
		parkingCfg: parkingCfg,
	}
}

func (ss *Session) Start() {
	ss.mu.Lock()
	if ss.state != statePaused && ss.state != stateStopped {
		ss.mu.Unlock()
		return
	}
	ss.state = stateRunning
	go ss.startTimer()

	ss.mu.Unlock()
	go ss.scheduleCar()
}

func (ss *Session) Pause() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	for _, car := range ss.car {
		if car.Timer != nil {
			car.Timer.Pause()
		}
	}

	ss.state = statePaused

	ss.timer.ticker.Stop()
}

func (ss *Session) Resume() {
	ss.mu.Lock()

	if ss.state != statePaused {
		ss.mu.Unlock()
		return
	}

	ss.state = stateRunning

	for _, car := range ss.car {
		if car.Timer != nil {
			car.Timer.Start()
		}
	}

	go ss.startTimer()

	ss.mu.Unlock()

	go ss.scheduleCar()
}

func (ss *Session) Stop() {
	ss.mu.Lock()

	if ss.state != stateStopped {
		ss.state = stateStopped
		ss.cancel()
	}

	for _, car := range ss.car {
		if car.Timer != nil {
			car.Timer.Stop()
		}
	}
	ss.mu.Unlock()
}

func (ss *Session) CheckPark(msg string) {
	if ss.state == stateRunning {
		args := strings.Split(msg, "park ")
		for _, carID := range args {
			if err := uuid.Validate(carID); err == nil {
				ss.sendParkEvent(carID)
			}
		}
	}
}

func (ss *Session) isRunning() bool {
	return ss.state == stateRunning && ss.ctx.Err() == nil
}

func (ss *Session) startTimer() {
	ss.timer.ticker = time.NewTicker(1 * time.Second)
	for {
		select {
		case _, ok := <-ss.timer.ticker.C:
			if !ok {
				return
			}
			ss.mu.Lock()
			if ss.state == stateRunning {
				ss.timer.elapsedTime = ss.timer.elapsedTime.Add(1 * time.Minute)
			}
			ss.mu.Unlock()
		case <-ss.ctx.Done():
			return
		}
	}
}
