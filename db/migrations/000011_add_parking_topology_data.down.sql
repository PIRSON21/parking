-- Откат изменений: обнуляем parking_topology
UPDATE parkings
SET parking_topology = NULL
WHERE parking_id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10);