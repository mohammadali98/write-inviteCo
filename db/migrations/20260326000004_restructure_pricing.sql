-- migrate:up

-- Remove old price columns
ALTER TABLE cards DROP COLUMN IF EXISTS price_pkr;
ALTER TABLE cards DROP COLUMN IF EXISTS price_nok;

-- Add new pricing columns
ALTER TABLE cards ADD COLUMN price_foil_pkr BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN price_nofoil_pkr BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN price_foil_nok BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN price_nofoil_nok BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN insert_price_pkr BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN insert_price_nok BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN min_order INT NOT NULL DEFAULT 1;
ALTER TABLE cards ADD COLUMN included_inserts INT NOT NULL DEFAULT 2;

-- Remove designs 4, 6, 11 (duplicates per client review)
DELETE FROM card_images WHERE card_id IN (SELECT id FROM cards WHERE name IN ('Blush Mehendi Wax Seal Suite', 'Emerald Nikkah Dua Suite', 'Mughal Garden Invitation'));
DELETE FROM cards WHERE name IN ('Blush Mehendi Wax Seal Suite', 'Emerald Nikkah Dua Suite', 'Mughal Garden Invitation');

-- Update pricing for remaining designs
-- Design 1: Vibrant Dholki Booklet (flat price, no foil option, min 3)
UPDATE cards SET price_foil_pkr = 1500, price_nofoil_pkr = 1500, price_foil_nok = 200, price_nofoil_nok = 200, insert_price_pkr = 0, insert_price_nok = 0, min_order = 3, included_inserts = 0 WHERE name = 'Vibrant Dholki Booklet';

-- Design 2: Sage & Gold Nikkah Suite
UPDATE cards SET price_foil_pkr = 580, price_nofoil_pkr = 530, price_foil_nok = 80, price_nofoil_nok = 70, insert_price_pkr = 80, insert_price_nok = 10, min_order = 50, included_inserts = 2 WHERE name = 'Sage & Gold Nikkah Suite';

-- Design 3: Blush Mehendi Laser-Cut
UPDATE cards SET price_foil_pkr = 580, price_nofoil_pkr = 530, price_foil_nok = 80, price_nofoil_nok = 70, insert_price_pkr = 80, insert_price_nok = 10, min_order = 50, included_inserts = 2 WHERE name = 'Blush Mehendi Laser-Cut';

-- Design 5: Sage Monogram Envelope
UPDATE cards SET price_foil_pkr = 380, price_nofoil_pkr = 350, price_foil_nok = 50, price_nofoil_nok = 45, insert_price_pkr = 50, insert_price_nok = 7, min_order = 50, included_inserts = 2 WHERE name = 'Sage Monogram Envelope';

-- Design 7: Royal Blue Silver Crest
UPDATE cards SET price_foil_pkr = 550, price_nofoil_pkr = 500, price_foil_nok = 75, price_nofoil_nok = 65, insert_price_pkr = 50, insert_price_nok = 7, min_order = 50, included_inserts = 2 WHERE name = 'Royal Blue Silver Crest';

-- Design 8: Emerald Baraat Suite
UPDATE cards SET price_foil_pkr = 380, price_nofoil_pkr = 350, price_foil_nok = 50, price_nofoil_nok = 45, insert_price_pkr = 50, insert_price_nok = 7, min_order = 50, included_inserts = 2 WHERE name = 'Emerald Baraat Suite';

-- Design 9: Burgundy Shalima Elegance
UPDATE cards SET price_foil_pkr = 480, price_nofoil_pkr = 430, price_foil_nok = 65, price_nofoil_nok = 55, insert_price_pkr = 80, insert_price_nok = 10, min_order = 50, included_inserts = 2 WHERE name = 'Burgundy Shalima Elegance';

-- Design 10: Floral Vellum Transparency
UPDATE cards SET price_foil_pkr = 380, price_nofoil_pkr = 350, price_foil_nok = 50, price_nofoil_nok = 45, insert_price_pkr = 50, insert_price_nok = 7, min_order = 50, included_inserts = 2 WHERE name = 'Floral Vellum Transparency';

-- Design 12: Burgundy Arch Shalima
UPDATE cards SET price_foil_pkr = 380, price_nofoil_pkr = 350, price_foil_nok = 50, price_nofoil_nok = 45, insert_price_pkr = 70, insert_price_nok = 9, min_order = 50, included_inserts = 2 WHERE name = 'Burgundy Arch Shalima';

-- Design 13: Mughal Rose Wax Seal Suite
UPDATE cards SET price_foil_pkr = 650, price_nofoil_pkr = 600, price_foil_nok = 85, price_nofoil_nok = 80, insert_price_pkr = 120, insert_price_nok = 15, min_order = 50, included_inserts = 2 WHERE name = 'Mughal Rose Wax Seal Suite';

-- Design 14: Forest Green Oval Walima
UPDATE cards SET price_foil_pkr = 580, price_nofoil_pkr = 530, price_foil_nok = 80, price_nofoil_nok = 70, insert_price_pkr = 80, insert_price_nok = 10, min_order = 50, included_inserts = 2 WHERE name = 'Forest Green Oval Walima';

-- Bid boxes and potlis (flat price, no foil/insert options)
UPDATE cards SET price_foil_pkr = 450, price_nofoil_pkr = 450, price_foil_nok = 70, price_nofoil_nok = 70, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Ivory Floral Acrylic Bid Box';
UPDATE cards SET price_foil_pkr = 450, price_nofoil_pkr = 450, price_foil_nok = 70, price_nofoil_nok = 70, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Ruby Gold Acrylic Bid Box';
UPDATE cards SET price_foil_pkr = 450, price_nofoil_pkr = 450, price_foil_nok = 70, price_nofoil_nok = 70, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Rose Quartz Ribbon Bid Box';
UPDATE cards SET price_foil_pkr = 350, price_nofoil_pkr = 350, price_foil_nok = 55, price_nofoil_nok = 55, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Peach Festive Potli Set';
UPDATE cards SET price_foil_pkr = 350, price_nofoil_pkr = 350, price_foil_nok = 55, price_nofoil_nok = 55, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Sunset Organza Potli Bundle';
UPDATE cards SET price_foil_pkr = 300, price_nofoil_pkr = 300, price_foil_nok = 50, price_nofoil_nok = 50, insert_price_pkr = 0, insert_price_nok = 0, min_order = 1, included_inserts = 0 WHERE name = 'Lavender Floral Favor Tins';

-- migrate:down
DELETE FROM cards WHERE name IN ('Blush Mehendi Wax Seal Suite', 'Emerald Nikkah Dua Suite', 'Mughal Garden Invitation');
ALTER TABLE cards DROP COLUMN price_foil_pkr;
ALTER TABLE cards DROP COLUMN price_nofoil_pkr;
ALTER TABLE cards DROP COLUMN price_foil_nok;
ALTER TABLE cards DROP COLUMN price_nofoil_nok;
ALTER TABLE cards DROP COLUMN insert_price_pkr;
ALTER TABLE cards DROP COLUMN insert_price_nok;
ALTER TABLE cards DROP COLUMN min_order;
ALTER TABLE cards DROP COLUMN included_inserts;
ALTER TABLE cards ADD COLUMN price_pkr BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN price_nok BIGINT NOT NULL DEFAULT 0;
