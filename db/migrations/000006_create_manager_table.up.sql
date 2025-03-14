CREATE TABLE IF NOT EXISTS manager(
    manager_id SERIAL PRIMARY KEY,
    manager_login CHAR(8) NOT NULL UNIQUE,
    manager_password TEXT NOT NULL,
    manager_email CHAR(15) NOT NULL UNIQUE
);