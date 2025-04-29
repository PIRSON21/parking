BEGIN;

-- Связываем парковки с менеджерами
UPDATE parkings SET manager_id = 1 WHERE parking_id IN (1, 2, 3); -- Первый менеджер рулит центром
UPDATE parkings SET manager_id = 2 WHERE parking_id IN (4, 5);    -- Второй менеджер рулит ТЦ
UPDATE parkings SET manager_id = 3 WHERE parking_id IN (6, 7);    -- Третий менеджер рулит ЖД
UPDATE parkings SET manager_id = 4 WHERE parking_id IN (8, 9, 10);-- Четвёртый менеджер рулит отелями и БЦ

COMMIT;