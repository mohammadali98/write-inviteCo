-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS extra_inserts_per_card BIGINT NOT NULL DEFAULT 0;

-- migrate:down
-- Forward-only migration. Keep existing order detail data intact.
