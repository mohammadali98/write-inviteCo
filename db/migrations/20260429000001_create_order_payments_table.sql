-- migrate:up
CREATE TABLE order_payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    payment_method TEXT NOT NULL DEFAULT 'bank_transfer',
    payment_status TEXT NOT NULL DEFAULT 'pending_payment',
    expected_amount BIGINT NOT NULL,
    submitted_amount BIGINT,
    sender_name TEXT,
    transaction_reference TEXT,
    proof_file_path TEXT,
    customer_note TEXT,
    submitted_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    admin_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_order_payments_method
        CHECK (payment_method IN ('bank_transfer')),
    CONSTRAINT chk_order_payments_status
        CHECK (payment_status IN ('pending_payment', 'awaiting_verification', 'payment_verified', 'payment_rejected')),
    CONSTRAINT chk_order_payments_expected_amount
        CHECK (expected_amount >= 0),
    CONSTRAINT chk_order_payments_submitted_amount
        CHECK (submitted_amount IS NULL OR submitted_amount >= 0)
);

CREATE INDEX idx_order_payments_status ON order_payments(payment_status);

INSERT INTO order_payments (
    order_id,
    payment_method,
    payment_status,
    expected_amount,
    verified_at,
    rejected_at,
    created_at,
    updated_at
)
SELECT
    o.id,
    'bank_transfer',
    CASE
        WHEN o.status IN ('confirmed', 'completed') THEN 'payment_verified'
        WHEN o.status = 'cancelled' THEN 'payment_rejected'
        ELSE 'pending_payment'
    END,
    o.total_price,
    CASE
        WHEN o.status IN ('confirmed', 'completed') THEN COALESCE(o.updated_at, o.created_at, CURRENT_TIMESTAMP)
        ELSE NULL
    END,
    CASE
        WHEN o.status = 'cancelled' THEN COALESCE(o.updated_at, o.created_at, CURRENT_TIMESTAMP)
        ELSE NULL
    END,
    COALESCE(o.created_at, CURRENT_TIMESTAMP),
    COALESCE(o.updated_at, o.created_at, CURRENT_TIMESTAMP)
FROM orders o
WHERE NOT EXISTS (
    SELECT 1
    FROM order_payments op
    WHERE op.order_id = o.id
);

-- migrate:down
DROP INDEX IF EXISTS idx_order_payments_status;
DROP TABLE IF EXISTS order_payments;
