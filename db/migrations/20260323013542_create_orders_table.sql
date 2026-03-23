-- migrate:up
CREATE TABLE orders
(
    id          BIGSERIAL PRIMARY KEY,
    customer_id BIGINT REFERENCES customers (id),
    card_id     BIGINT REFERENCES cards (id),
    quantity    BIGINT      NOT NULL DEFAULT 1,
    total_price BIGINT      NOT NULL,
    status      TEXT                 DEFAULT 'pending',
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_customer_id ON orders (customer_id);
CREATE INDEX idx_orders_card_id ON orders (card_id);
CREATE INDEX idx_orders_status ON orders (status);

-- migrate:down
DROP INDEX idx_orders_status;
DROP INDEX idx_orders_card_id;
DROP INDEX idx_orders_customer_id;
DROP TABLE orders;

