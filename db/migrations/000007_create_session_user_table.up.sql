CREATE TABLE IF NOT EXISTS user_session (
    session_id UUID UNIQUE PRIMARY KEY,
    user_id INTEGER NULL,
    deadline TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES manager(manager_id) ON DELETE CASCADE
);