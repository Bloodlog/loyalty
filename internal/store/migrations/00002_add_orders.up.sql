BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_number BIGINT NOT NULL UNIQUE,
    status_id INT NOT NULL,
    user_id INT NOT NULL,
    accrual FLOAT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
);

COMMIT;