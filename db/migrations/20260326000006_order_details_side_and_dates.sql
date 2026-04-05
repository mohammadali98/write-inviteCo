-- migrate:up
ALTER TABLE order_details
    ADD COLUMN side TEXT NOT NULL DEFAULT 'bride',
    ADD COLUMN mehndi_date DATE,
    ADD COLUMN baraat_date DATE,
    ADD COLUMN nikkah_date DATE,
    ADD COLUMN walima_date DATE;

ALTER TABLE order_details
    ADD CONSTRAINT chk_order_details_side CHECK (side IN ('bride', 'groom'));

ALTER TABLE order_details ALTER COLUMN bride_name DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN groom_name DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN bride_father_name DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN groom_father_name DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_type DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_date DROP NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_time DROP NOT NULL;

-- migrate:down
ALTER TABLE order_details ALTER COLUMN bride_name SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN groom_name SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN bride_father_name SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN groom_father_name SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_type SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_date SET NOT NULL;
ALTER TABLE order_details ALTER COLUMN event_time SET NOT NULL;

ALTER TABLE order_details DROP CONSTRAINT chk_order_details_side;

ALTER TABLE order_details
    DROP COLUMN walima_date,
    DROP COLUMN nikkah_date,
    DROP COLUMN baraat_date,
    DROP COLUMN mehndi_date,
    DROP COLUMN side;
