\c gophermart;

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_number BIGINT NOT NULL UNIQUE,
    status_id INT NOT NULL,
    user_id INT NOT NULL,
    accrual FLOAT NULL,
    next_attempt TIMESTAMP NULL,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
);

GRANT ALL PRIVILEGES ON TABLE orders TO gopher;