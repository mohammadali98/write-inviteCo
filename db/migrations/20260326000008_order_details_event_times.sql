-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS mehndi_time_type TEXT,
    ADD COLUMN IF NOT EXISTS mehndi_time TIME,
    ADD COLUMN IF NOT EXISTS baraat_time_type TEXT,
    ADD COLUMN IF NOT EXISTS baraat_time TIME,
    ADD COLUMN IF NOT EXISTS baraat_arrival_time TIME,
    ADD COLUMN IF NOT EXISTS rukhsati_time TIME,
    ADD COLUMN IF NOT EXISTS nikkah_time_type TEXT,
    ADD COLUMN IF NOT EXISTS nikkah_time TIME,
    ADD COLUMN IF NOT EXISTS walima_time_type TEXT,
    ADD COLUMN IF NOT EXISTS walima_time TIME,
    ADD COLUMN IF NOT EXISTS reception_time TIME,
    ADD COLUMN IF NOT EXISTS dinner_time TIME;

ALTER TABLE order_details
    ALTER COLUMN mehndi_dinner_time TYPE TIME USING NULLIF(mehndi_dinner_time::text, '')::time,
    ALTER COLUMN baraat_dinner_time TYPE TIME USING NULLIF(baraat_dinner_time::text, '')::time,
    ALTER COLUMN nikkah_dinner_time TYPE TIME USING NULLIF(nikkah_dinner_time::text, '')::time,
    ALTER COLUMN walima_dinner_time TYPE TIME USING NULLIF(walima_dinner_time::text, '')::time;

-- migrate:down
ALTER TABLE order_details
    ALTER COLUMN mehndi_dinner_time TYPE TEXT USING CASE WHEN mehndi_dinner_time IS NULL THEN NULL ELSE mehndi_dinner_time::text END,
    ALTER COLUMN baraat_dinner_time TYPE TEXT USING CASE WHEN baraat_dinner_time IS NULL THEN NULL ELSE baraat_dinner_time::text END,
    ALTER COLUMN nikkah_dinner_time TYPE TEXT USING CASE WHEN nikkah_dinner_time IS NULL THEN NULL ELSE nikkah_dinner_time::text END,
    ALTER COLUMN walima_dinner_time TYPE TEXT USING CASE WHEN walima_dinner_time IS NULL THEN NULL ELSE walima_dinner_time::text END;

ALTER TABLE order_details
    DROP COLUMN IF EXISTS dinner_time,
    DROP COLUMN IF EXISTS reception_time,
    DROP COLUMN IF EXISTS walima_time,
    DROP COLUMN IF EXISTS walima_time_type,
    DROP COLUMN IF EXISTS nikkah_time,
    DROP COLUMN IF EXISTS nikkah_time_type,
    DROP COLUMN IF EXISTS rukhsati_time,
    DROP COLUMN IF EXISTS baraat_arrival_time,
    DROP COLUMN IF EXISTS baraat_time,
    DROP COLUMN IF EXISTS baraat_time_type,
    DROP COLUMN IF EXISTS mehndi_time,
    DROP COLUMN IF EXISTS mehndi_time_type;
