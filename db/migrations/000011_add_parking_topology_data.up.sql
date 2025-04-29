-- Добавляем данные о топологии имеющимся парковкам.

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", "I"],
    [".", "P", "P", "P", "."],
    [".", "D", "D", ".", "."],
    [".", ".", ".", ".", "."],
    ["O", ".", ".", "P", "P"]
]'
WHERE parking_id = 1;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", ".", "I"],
    [".", "P", "P", "P", "P", "."],
    [".", "D", "D", ".", ".", "."],
    [".", ".", ".", ".", ".", "."],
    ["O", ".", ".", "P", "P", "P"]
]'
WHERE parking_id = 2;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", "."],
    [".", "P", "P", ".", "."],
    [".", "D", ".", ".", "."],
    [".", ".", ".", ".", "."]
]'
WHERE parking_id = 3;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", "."],
    [".", "P", "P", "P", "."],
    [".", "D", "D", ".", "."],
    [".", ".", ".", ".", "."],
    ["O", ".", ".", ".", "I"]
]'
WHERE parking_id = 4;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", ".", "I"],
    [".", "P", "P", "P", ".", "."],
    [".", "D", ".", ".", ".", "."],
    [".", ".", ".", ".", ".", "."],
    ["O", ".", ".", "P", "P", "P"]
]'
WHERE parking_id = 5;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", "."],
    [".", "P", "P", "."],
    [".", "D", ".", "."],
    [".", ".", ".", "."]
]'
WHERE parking_id = 6;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", ".", "I"],
    [".", "P", "P", "P", "P", "."],
    [".", "D", ".", ".", ".", "."],
    [".", ".", ".", ".", ".", "."],
    ["O", ".", ".", ".", "P", "P"]
]'
WHERE parking_id = 7;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", ".", "I"],
    [".", "P", "P", "P", ".", "."],
    [".", "D", ".", ".", ".", "."],
    [".", ".", ".", ".", ".", "."],
    ["O", ".", ".", "P", "P", "."]
]'
WHERE parking_id = 8;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", ".", "."],
    [".", "P", "P", "P", "P", "I"],
    [".", "D", ".", ".", ".", "."],
    [".", ".", ".", ".", ".", "."],
    ["O", ".", ".", "P", "P", "P"]
]'
WHERE parking_id = 9;

UPDATE parkings
SET parking_topology = '[
    [".", ".", ".", ".", "."],
    [".", "P", "P", ".", "."],
    [".", "D", ".", ".", "."],
    [".", ".", ".", ".", "."],
    ["O", ".", ".", ".", "I"]
]'
WHERE parking_id = 10;
