CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(255) NOT NULL,
    sum DOUBLE PRECISION NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    processed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
