CREATE TABLE IF NOT EXISTS parking_cell (
    parking_id INT NOT NULL REFERENCES parkings(parking_id) ON DELETE CASCADE,
    x INT NOT NULL,
    y INT NOT NULL,
    cell_type VARCHAR(10) NOT NULL CHECK (cell_type IN ('P', 'D', 'I', 'O')),
    UNIQUE (parking_id, x, y)
);

CREATE UNIQUE INDEX idx_parking_cell_unique ON parking_cell (parking_id, x, y);