-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS mehndi_day TEXT,
    ADD COLUMN IF NOT EXISTS baraat_day TEXT,
    ADD COLUMN IF NOT EXISTS nikkah_day TEXT,
    ADD COLUMN IF NOT EXISTS walima_day TEXT;

ALTER TABLE order_details
    DROP COLUMN IF EXISTS dinner_time;

-- migrate:down
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS dinner_time TIME;

ALTER TABLE order_details
    DROP COLUMN IF EXISTS mehndi_day,
    DROP COLUMN IF EXISTS baraat_day,
    DROP COLUMN IF EXISTS nikkah_day,
    DROP COLUMN IF EXISTS walima_day;
