CREATE TABLE IF NOT EXISTS manager(
    manager_id SERIAL PRIMARY KEY,
    manager_login VARCHAR(20) NOT NULL,
    manager_password TEXT NOT NULL,
    manager_email VARCHAR(15) NOT NULL UNIQUE
);
