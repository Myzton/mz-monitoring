CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL
    );

CREATE TABLE IF NOT EXISTS targets (
      id SERIAL PRIMARY KEY,
       user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    interval_sec INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
    );

CREATE TABLE IF NOT EXISTS check_logs (
     id BIGSERIAL PRIMARY KEY,
    target_id INT NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    status_code INT NOT NULL,
    response_time_ms INT NOT NULL,
    is_up BOOLEAN NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW()
    );