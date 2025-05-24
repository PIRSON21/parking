UPDATE parkings
SET parking_topology = '[[".", ".", ".", ".", "."], ["D", ".", ".", ".", "P"], [".", "P", ".", ".", "."], [".", "I", ".", ".", "O"]]'
WHERE parking_id = 1;

UPDATE parkings
SET parking_topology = '[[".", ".", ".", ".", "P"], [".", ".", ".", ".", "P"], [".", "P", ".", ".", "."], [".", ".", "D", "P", "P"], ["P", ".", "D", ".", "D"], [".", "I", "O", ".", "."]]'
WHERE parking_id = 2;

UPDATE parkings
SET parking_topology = '[[".", ".", ".", "."], [".", "P", ".", "."], [".", "P", "P", "."], ["P", "P", ".", "."], ["O", ".", ".", "I"]]'
WHERE parking_id = 3;

UPDATE parkings
SET parking_topology = '[["P", ".", "D", ".", "P"], [".", ".", ".", "P", "P"], ["P", ".", ".", "P", "P"], [".", ".", ".", ".", "."], ["O", ".", ".", ".", "I"]]'
WHERE parking_id = 4;

UPDATE parkings
SET parking_topology = '[["D", "P", "P", ".", "."], [".", ".", "D", ".", "P"], [".", ".", "P", "P", "."], [".", ".", ".", ".", "."], ["P", "P", ".", "P", "."], [".", "O", "I", ".", "."]]'
WHERE parking_id = 5;

UPDATE parkings
SET parking_topology = '[[".", ".", ".", "P"], [".", ".", ".", "P"], ["P", "P", "P", "."], [".", "O", ".", "I"]]'
WHERE parking_id = 6;

UPDATE parkings
SET parking_topology = '[[".", "D", ".", "D"], ["P", "P", ".", "."], [".", "D", "D", "."], [".", ".", "D", "."], [".", ".", ".", "P"], [".", "O", ".", "I"]]'
WHERE parking_id = 7;

UPDATE parkings
SET parking_topology = '[["P", "P", ".", "."], [".", ".", ".", "P"], [".", "D", "P", "."], [".", "P", ".", "."], [".", "P", "P", "."], ["O", ".", ".", "I"]]'
WHERE parking_id = 8;

UPDATE parkings
SET parking_topology = '[["P", "P", ".", ".", "D", "."], ["D", "P", ".", ".", ".", "P"], [".", ".", ".", "P", "D", "P"], ["P", ".", "D", "P", ".", "."], ["I", "O", ".", ".", ".", "."]]'
WHERE parking_id = 9;

UPDATE parkings
SET parking_topology = '[["P", ".", ".", "P", "P"], [".", "P", ".", ".", "P"], ["D", ".", "P", "P", "."], [".", ".", "P", ".", "."], [".", ".", "I", "O", "."]]'
WHERE parking_id = 10;
