-- migrate:up
ALTER TABLE order_details
    ADD COLUMN mehndi_dinner_time TEXT,
    ADD COLUMN baraat_dinner_time TEXT,
    ADD COLUMN nikkah_dinner_time TEXT,
    ADD COLUMN walima_dinner_time TEXT;

-- migrate:down
ALTER TABLE order_details
    DROP COLUMN walima_dinner_time,
    DROP COLUMN nikkah_dinner_time,
    DROP COLUMN baraat_dinner_time,
    DROP COLUMN mehndi_dinner_time;
