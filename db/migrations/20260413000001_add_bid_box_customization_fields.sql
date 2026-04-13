-- migrate:up
ALTER TABLE order_details
    ADD COLUMN IF NOT EXISTS top_label TEXT,
    ADD COLUMN IF NOT EXISTS couple_name TEXT,
    ADD COLUMN IF NOT EXISTS bid_box_details TEXT;

-- migrate:down
ALTER TABLE order_details DROP COLUMN IF EXISTS bid_box_details;
ALTER TABLE order_details DROP COLUMN IF EXISTS couple_name;
ALTER TABLE order_details DROP COLUMN IF EXISTS top_label;
