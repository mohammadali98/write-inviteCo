-- migrate:up
CREATE TABLE customers
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT,
    phone      TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_customers_email ON customers (email);

-- migrate:down
DROP INDEX idx_customers_email;
DROP TABLE customers;
