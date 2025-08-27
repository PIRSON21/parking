package models

// item - структура для реализации очереди с приоритетом.
type item struct {
	value float64 // расстояние
	x, y  int     // координаты
	index int     // индекс в куче
}

// priorityQueue - очередь с приоритетом для алгоритма Дейкстры.
type priorityQueue []item

// Len возвращает длину очереди.
func (pq priorityQueue) Len() int { return len(pq) }

// Less сравнивает два элемента в очереди с приоритетом.
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

// Swap меняет местами два элемента в очереди с приоритетом.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push добавляет элемент в очередь с приоритетом.
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop удаляет элемент с наименьшим приоритетом из очереди.
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // для безопасности
	*pq = old[0 : n-1]
	return item
}
