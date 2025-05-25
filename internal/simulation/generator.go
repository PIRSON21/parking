package simulation

import (
	"math"
	"math/rand/v2"
	"time"
)

const (
	dayStartHour   = 6
	nightStartHour = 22
)

// generateArrivalDelay вычисляет задержку появления автомобиля в зависимости от типа потока.
func (ss *Session) generateArrivalDelay() time.Duration {
	switch ss.arrivalCfg.Type {
	case "exponential":
		return generateExponentialDelay(ss.arrivalCfg.Lambda)
	case "normal":
		return generateNormalDelay(ss.arrivalCfg.Mean, ss.arrivalCfg.StdDev)
	case "uniform":
		return generateUniformDelay(ss.arrivalCfg.MinDelay, ss.arrivalCfg.MaxDelay)
	case "discrete":
		return generateDiscreteDelay(ss.arrivalCfg.DiscreteTime)
	default:
		return generateDiscreteDelay(ss.arrivalCfg.DiscreteTime)
	}
}

// generateExponentialDelay вычисляет задержку экспоненциального распределения.
func generateExponentialDelay(lambda float64) time.Duration {
	r := rand.Float64()
	delay := -math.Log(1.0-r) / lambda
	return time.Duration(delay * float64(time.Second))
}

// generateNormalDelay вычисляет задержку нормального распределения.
func generateNormalDelay(mean float64, dev float64) time.Duration {
	delay := rand.NormFloat64()*dev + mean
	return time.Duration(math.Abs(delay) * float64(time.Second))
}

// generateUniformDelay вычисляет задержку равномерного распределения.
func generateUniformDelay(minDelay float64, maxDelay float64) time.Duration {
	delay := minDelay + (maxDelay-minDelay)*rand.Float64()
	return time.Duration(delay * float64(time.Second))
}

// generateDiscreteDelay вычисляет задержку дискретного потока.
func generateDiscreteDelay(discrete float64) time.Duration {
	return time.Duration(discrete * float64(time.Second))
}

func (ss *Session) evaluateEntrance() bool {
	return rand.Float64() < ss.arrivalCfg.ParkingProb
}

func (ss *Session) generateLeaveDelay() time.Duration {
	switch ss.parkingCfg.Type {
	case "exponential":
		return generateExponentialDelay(ss.parkingCfg.Lambda)
	case "normal":
		return generateNormalDelay(ss.parkingCfg.Mean, ss.parkingCfg.StdDev)
	case "uniform":
		return generateUniformDelay(ss.parkingCfg.MinDuration, ss.parkingCfg.MaxDuration)
	case "discrete":
		return generateDiscreteDelay(ss.parkingCfg.DiscreteTime)
	default:
		return generateDiscreteDelay(ss.parkingCfg.DiscreteTime)
	}
}

// calculateParkingCost вычисляет стоимость стоянки.
func (ss *Session) calculateParkingCost(now time.Time, entered time.Time) float64 {
	totalCost := 0.0

	for entered.Before(now) {
		nextBoundary := getNextBoundary(entered)

		// если машина уезжает до следующей смены тарифа
		if now.Before(nextBoundary) {
			duration := now.Sub(entered)
			totalCost += calculateSegmentCost(duration, entered, ss.parking.DayTariff, ss.parking.NightTariff)
			break
		}
		// если уезжает позже смены
		duration := nextBoundary.Sub(entered)
		totalCost += calculateSegmentCost(duration, entered, ss.parking.DayTariff, ss.parking.NightTariff)

		entered = nextBoundary
	}

	return totalCost
}

// getNextBoundary определяет следующую границу тарифа.
func getNextBoundary(t time.Time) time.Time {
	hour := t.Hour()

	if hour >= nightStartHour || hour < dayStartHour {
		return time.Date(t.Year(), t.Month(), t.Day()+1, dayStartHour, 0, 0, 0, t.Location())
	}

	return time.Date(t.Year(), t.Month(), t.Day(), nightStartHour, 0, 0, 0, t.Location())
}

// calculateSegmentCost вычисляет стоимость парковки для длительности d.
func calculateSegmentCost(d time.Duration, t time.Time, rateDay, rateNight float64) float64 {
	hour := t.Hour()

	if hour >= nightStartHour || hour < dayStartHour {
		return d.Hours() * rateNight
	}

	return d.Hours() * rateDay
}
