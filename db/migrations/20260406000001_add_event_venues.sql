-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS mehndi_venue_name TEXT,
    ADD COLUMN IF NOT EXISTS mehndi_venue_address TEXT,
    ADD COLUMN IF NOT EXISTS baraat_venue_name TEXT,
    ADD COLUMN IF NOT EXISTS baraat_venue_address TEXT,
    ADD COLUMN IF NOT EXISTS nikkah_venue_name TEXT,
    ADD COLUMN IF NOT EXISTS nikkah_venue_address TEXT,
    ADD COLUMN IF NOT EXISTS walima_venue_name TEXT,
    ADD COLUMN IF NOT EXISTS walima_venue_address TEXT;

UPDATE order_details
SET
    mehndi_venue_name = COALESCE(mehndi_venue_name, venue_name),
    mehndi_venue_address = COALESCE(mehndi_venue_address, venue_address),
    baraat_venue_name = COALESCE(baraat_venue_name, venue_name),
    baraat_venue_address = COALESCE(baraat_venue_address, venue_address),
    nikkah_venue_name = COALESCE(nikkah_venue_name, venue_name),
    nikkah_venue_address = COALESCE(nikkah_venue_address, venue_address),
    walima_venue_name = COALESCE(walima_venue_name, venue_name),
    walima_venue_address = COALESCE(walima_venue_address, venue_address);

ALTER TABLE order_details
    DROP COLUMN IF EXISTS venue_name,
    DROP COLUMN IF EXISTS venue_address;

-- migrate:down
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS venue_name TEXT,
    ADD COLUMN IF NOT EXISTS venue_address TEXT;

UPDATE order_details
SET
    venue_name = COALESCE(
        NULLIF(mehndi_venue_name, ''),
        NULLIF(baraat_venue_name, ''),
        NULLIF(nikkah_venue_name, ''),
        NULLIF(walima_venue_name, '')
    ),
    venue_address = COALESCE(
        NULLIF(mehndi_venue_address, ''),
        NULLIF(baraat_venue_address, ''),
        NULLIF(nikkah_venue_address, ''),
        NULLIF(walima_venue_address, '')
    );

ALTER TABLE order_details
    DROP COLUMN IF EXISTS mehndi_venue_name,
    DROP COLUMN IF EXISTS mehndi_venue_address,
    DROP COLUMN IF EXISTS baraat_venue_name,
    DROP COLUMN IF EXISTS baraat_venue_address,
    DROP COLUMN IF EXISTS nikkah_venue_name,
    DROP COLUMN IF EXISTS nikkah_venue_address,
    DROP COLUMN IF EXISTS walima_venue_name,
    DROP COLUMN IF EXISTS walima_venue_address;
