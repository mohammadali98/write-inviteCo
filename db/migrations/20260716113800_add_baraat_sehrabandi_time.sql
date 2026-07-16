-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS baraat_sehrabandi_time TIME;

-- migrate:down
ALTER TABLE order_details
    DROP COLUMN IF EXISTS baraat_sehrabandi_time;
