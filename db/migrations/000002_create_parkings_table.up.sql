CREATE TABLE IF NOT EXISTS parkings (
    parking_id SERIAL PRIMARY KEY,
    parking_name VARCHAR(10) NOT NULL,
    parking_address VARCHAR(30) NOT NULL,
    parking_width INT,
    parking_height INT
);

CREATE INDEX idx_parkings_parking_name ON parkings(parking_name);