-- migrate:up
ALTER TABLE products
    ADD COLUMN card_id BIGINT REFERENCES cards(id) ON DELETE SET NULL;

CREATE INDEX idx_products_card_id ON products(card_id);

-- migrate:down
DROP INDEX IF EXISTS idx_products_card_id;

ALTER TABLE products
    DROP COLUMN IF EXISTS card_id;
