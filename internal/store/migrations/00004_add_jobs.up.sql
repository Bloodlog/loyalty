BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS jobs (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_id BIGINT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    pool_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT fk_jobs_order FOREIGN KEY (order_id) REFERENCES orders(id)
);

COMMIT;