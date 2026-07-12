-- migrate:up
ALTER TABLE orders ADD COLUMN public_token UUID NOT NULL DEFAULT gen_random_uuid();
CREATE UNIQUE INDEX idx_orders_public_token ON orders (public_token);

-- migrate:down
DROP INDEX idx_orders_public_token;
ALTER TABLE orders DROP COLUMN public_token;
