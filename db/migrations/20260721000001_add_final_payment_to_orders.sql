-- migrate:up
ALTER TABLE orders
    ADD COLUMN final_payment_status TEXT NOT NULL DEFAULT 'pending_payment',
    ADD COLUMN final_payment_proof_url TEXT,
    ADD COLUMN final_payment_sender_name TEXT,
    ADD COLUMN final_payment_submitted_at TIMESTAMPTZ,
    ADD COLUMN final_payment_verified_at TIMESTAMPTZ,
    ADD COLUMN final_payment_rejected_at TIMESTAMPTZ,
    ADD COLUMN final_payment_admin_note TEXT;

CREATE INDEX idx_orders_final_payment_status ON orders(final_payment_status);

-- migrate:down
DROP INDEX idx_orders_final_payment_status;
ALTER TABLE orders
    DROP COLUMN final_payment_status,
    DROP COLUMN final_payment_proof_url,
    DROP COLUMN final_payment_sender_name,
    DROP COLUMN final_payment_submitted_at,
    DROP COLUMN final_payment_verified_at,
    DROP COLUMN final_payment_rejected_at,
    DROP COLUMN final_payment_admin_note;
