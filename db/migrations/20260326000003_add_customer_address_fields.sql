-- migrate:up
ALTER TABLE customers ADD COLUMN address TEXT;
ALTER TABLE customers ADD COLUMN city TEXT;
ALTER TABLE customers ADD COLUMN postal_code TEXT;

-- migrate:down
ALTER TABLE customers DROP COLUMN address;
ALTER TABLE customers DROP COLUMN city;
ALTER TABLE customers DROP COLUMN postal_code;
