package models

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"github.com/ivahaev/timer"
)

// ParkingSpot отражает парковочное место.
type ParkingPoint struct {
	cell   ParkingCell
	X, Y   int
	isFree bool
}

// ParkingLot отражает топологию парковки.
type ParkingLot struct {
	mu          sync.Mutex
	topology    [][]*ParkingPoint
	EntryX      int
	EntryY      int
	DayTariff   float64
	NightTariff float64
}

// NewParkingLot создает модель парковки для сессии.
func NewParkingLot(parking *Parking) *ParkingLot {
	var entryX, entryY int

	var topology [][]*ParkingPoint

	for width := 0; width < len(parking.Cells); width++ {
		topology = append(topology, []*ParkingPoint{})
		for height := 0; height < len(parking.Cells[width]); height++ {
			topology[width] = append(topology[width], &ParkingPoint{
				cell:   parking.Cells[width][height],
				X:      width,
				Y:      height,
				isFree: true,
			})
			if parking.Cells[width][height].IsEntrance() {
				entryX = width
				entryY = height
			}
		}
	}

	return &ParkingLot{
		topology:    topology,
		EntryX:      entryX,
		EntryY:      entryY,
		DayTariff:   float64(*parking.DayTariff),
		NightTariff: float64(*parking.NightTariff),
	}
}

// HasFreeSpot проверит парковку на свободные места.
func (p *ParkingLot) HasFreeSpot() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return false
}

// OccupySpot занимает парковочное место (если найдёт). Вернёт nil, false если места нет.
func (p *ParkingLot) OccupySpot() (*ParkingPoint, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var nearestSpot *ParkingPoint
	minDistance := math.MaxFloat64

	nearestSpot = p.findNearestSpot(minDistance)

	if nearestSpot != nil {
		nearestSpot.isFree = false
		return nearestSpot, true
	}

	return nil, false
}

// findNearestSpot находит ближайшее свободное парковочное место с помощью алгоритма Дейкстры.
func (p *ParkingLot) findNearestSpot(minDistance float64) *ParkingPoint {
	height := len(p.topology)
	if height == 0 {
		return nil
	}
	width := len(p.topology[0])

	// Направления движения (вверх, вправо, вниз, влево)
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}

	// Матрица расстояний
	dist := make([][]float64, height)
	for i := range dist {
		dist[i] = make([]float64, width)
		for j := range dist[i] {
			dist[i][j] = math.Inf(1)
		}
	}

	// Матрица для отслеживания посещенных ячеек
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	// Очередь с приоритетом для алгоритма Дейкстры
	queue := make(priorityQueue, 0)

	// Начинаем с точки входа
	dist[p.EntryX][p.EntryY] = 0
	queue = append(queue, item{value: 0, x: p.EntryX, y: p.EntryY})
	heap.Init(&queue)

	// Лучшее парковочное место
	var bestParkingSpot *ParkingPoint
	var bestDistance float64 = math.Inf(1)

	// Алгоритм Дейкстры
	for queue.Len() > 0 {
		current := heap.Pop(&queue).(item)
		x, y := current.x, current.y

		// Если уже посетили ячейку или расстояние больше, чем лучшее найденное, пропускаем
		if visited[x][y] || dist[x][y] > bestDistance {
			continue
		}
		visited[x][y] = true

		// Если это парковочное место и оно свободно
		if p.topology[x][y].cell.IsParking() && p.topology[x][y].isFree {
			if dist[x][y] < bestDistance {
				bestDistance = dist[x][y]
				bestParkingSpot = p.topology[x][y]
			}
		}

		for i := 0; i < 4; i++ {
			nx, ny := x+dx[i], y+dy[i]

			// Проверяем границы
			if nx < 0 || nx >= height || ny < 0 || ny >= width {
				continue
			}

			// Если клетка не дорога, не выезд, и не парковочное место
			cell := p.topology[nx][ny].cell
			if (!cell.IsRoad() && !cell.IsExit() && !cell.IsParking()) || visited[nx][ny] {
				// fmt.Printf("cell %s (%d, %d) is not road, exit or parking\n", cell, nx, ny)
				continue
			}

			// Вычисляем новое расстояние
			newDistance := dist[x][y] + 1

			// Если новое расстояние меньше, обновляем
			if newDistance < dist[nx][ny] {
				dist[nx][ny] = newDistance
				heap.Push(&queue, item{value: newDistance, x: nx, y: ny})
			}
		}
	}

	// Если нашли лучшее парковочное место, возвращаем его
	if bestParkingSpot != nil {
		return bestParkingSpot
	}

	return nil
}

// ReleaseSpot освобождает занятое парковочное место.
func (p *ParkingLot) ReleaseSpot(spot *ParkingPoint) {
	p.mu.Lock()
	defer p.mu.Unlock()

	spot.isFree = true
}

// SimulatedCar описывает машину в симуляции.
type SimulatedCar struct {
	CarID     string
	State     string
	Timer     *timer.Timer
	Spot      *ParkingPoint
	EnterTime time.Time
	Price     float64
}

// PathPoint представляет точку на пути
type PathPoint struct {
	X, Y int
}

// Path представляет путь от одной точки до другой
type Path struct {
	Points   []PathPoint
	Distance float64
	IsValid  bool
}

// FindPath находит путь от точки A до точки B на парковке (мок функция)
func (p *ParkingLot) FindPath(fromX, fromY, toX, toY int) *Path {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Проверяем границы
	height := len(p.topology)
	if height == 0 {
		return &Path{Points: nil, Distance: 0, IsValid: false}
	}
	width := len(p.topology[0])

	if fromX < 0 || fromX >= height || fromY < 0 || fromY >= width ||
		toX < 0 || toX >= height || toY < 0 || toY >= width {
		return &Path{Points: nil, Distance: 0, IsValid: false}
	}

	// Мок реализация - простой прямой путь
	path := &Path{
		Points:   make([]PathPoint, 0),
		Distance: 0,
		IsValid:  true,
	}

	// Добавляем начальную точку
	path.Points = append(path.Points, PathPoint{X: fromX, Y: fromY})

	// Простой алгоритм: сначала двигаемся по X, потом по Y
	currentX, currentY := fromX, fromY

	// Движение по оси X
	for currentX != toX {
		if currentX < toX {
			currentX++
		} else {
			currentX--
		}

		// Проверяем, что ячейка проходима
		cell := p.topology[currentX][currentY].cell
		if !cell.IsRoad() && !cell.IsExit() && !cell.IsParking() && !cell.IsEntrance() {
			return &Path{Points: nil, Distance: 0, IsValid: false}
		}

		path.Points = append(path.Points, PathPoint{X: currentX, Y: currentY})
		path.Distance++
	}

	// Движение по оси Y
	for currentY != toY {
		if currentY < toY {
			currentY++
		} else {
			currentY--
		}

		// Проверяем, что ячейка проходима
		cell := p.topology[currentX][currentY].cell
		if !cell.IsRoad() && !cell.IsExit() && !cell.IsParking() && !cell.IsEntrance() {
			return &Path{Points: nil, Distance: 0, IsValid: false}
		}

		path.Points = append(path.Points, PathPoint{X: currentX, Y: currentY})
		path.Distance++
	}

	return path
}

// GetPathToSpot находит путь от входа до указанного парковочного места
func (p *ParkingLot) GetPathToSpot(spot *ParkingPoint) *Path {
	return p.FindPath(p.EntryX, p.EntryY, spot.X, spot.Y)
}

// GetPathFromSpot находит путь от парковочного места до выхода
func (p *ParkingLot) GetPathFromSpot(spot *ParkingPoint) *Path {
	// Находим выход
	exitX, exitY := p.findExit()
	if exitX == -1 || exitY == -1 {
		// Если выхода нет, используем вход как выход
		exitX, exitY = p.EntryX, p.EntryY
	}

	return p.FindPath(spot.X, spot.Y, exitX, exitY)
}

// findExit находит координаты выхода
func (p *ParkingLot) findExit() (int, int) {
	for x := 0; x < len(p.topology); x++ {
		for y := 0; y < len(p.topology[x]); y++ {
			if p.topology[x][y].cell.IsExit() {
				return x, y
			}
		}
	}
	return -1, -1
}
