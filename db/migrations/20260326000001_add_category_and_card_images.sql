-- migrate:up

-- Add category column to cards
ALTER TABLE cards ADD COLUMN category TEXT NOT NULL DEFAULT 'wedding-cards';

-- Remove old price column, add dual currency
ALTER TABLE cards DROP COLUMN price;
ALTER TABLE cards ADD COLUMN price_pkr BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN price_nok BIGINT NOT NULL DEFAULT 0;

-- Create card_images table for multiple images per card
CREATE TABLE card_images (
    id          BIGSERIAL PRIMARY KEY,
    card_id     BIGINT      NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    image       TEXT        NOT NULL,
    sort_order  INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_card_images_card_id ON card_images(card_id);
CREATE INDEX idx_cards_category ON cards(category);

-- Add currency column to orders
ALTER TABLE orders ADD COLUMN currency TEXT NOT NULL DEFAULT 'PKR';

-- Seed the 14 wedding card designs
-- First delete the old sample data
DELETE FROM orders;
DELETE FROM cards;

INSERT INTO cards (name, description, price_pkr, price_nok, image, category) VALUES
('Vibrant Dholki Booklet', 'Colorful spiral-bound Dholki ceremony booklets with traditional illustrations, dhol players, and floral motifs in vibrant reds and pinks', 250, 40, '/static/cards/design-01/27DEBC49-9228-41D7-B621-873B77C8E73F.JPG', 'wedding-cards'),
('Sage & Gold Nikkah Suite', 'Sage green envelope suite with gold foil monogram, wax seal, and complete nikkah ceremony stationery with baby''s breath styling', 250, 40, '/static/cards/design-02/homepage1.JPG', 'wedding-cards'),
('Blush Mehendi Laser-Cut', 'Blush pink mehendi invitation with intricate laser-cut floral overlay, Urdu calligraphy, silver foil monogram, and save the date card', 250, 40, '/static/cards/design-03/homepage2.jpg', 'wedding-cards'),
('Blush Mehendi Wax Seal Suite', 'Elegant blush pink laser-cut mehendi invitation with embossed floral details, gold wax seals, and botanical illustrations on textured cream card', 250, 40, '/static/cards/design-04/IMG_0304.jpg', 'wedding-cards'),
('Sage Monogram Envelope', 'Sage green envelopes with custom gold foil monogram initials and gold leaf wax seal — perfect for a refined nikkah suite', 250, 40, '/static/cards/design-05/IMG_0449.jpg', 'wedding-cards'),
('Emerald Nikkah Dua Suite', 'Dark emerald green envelope suite with arch-shaped dua card, gold foil crest, calligraphy names, and wax seals with baby''s breath styling', 250, 40, '/static/cards/design-06/IMG_1679.jpg', 'wedding-cards'),
('Royal Blue Silver Crest', 'Deep royal blue invitation with silver foil floral crest monogram, Write&Invite branding, and white wax seal — bold and regal', 250, 40, '/static/cards/design-07/IMG_1721.jpg', 'wedding-cards'),
('Emerald Baraat Suite', 'Deep emerald green envelope with gold foil monogram, wax seal, paired with cream baraat and nikkah cards featuring sunflower illustrations and QR codes', 250, 40, '/static/cards/design-08/IMG_4429.jpg', 'wedding-cards'),
('Burgundy Shalima Elegance', 'Rich burgundy linen-textured envelopes with gold foil monogram, embossed cream card with vellum overlay and gold foil seal', 250, 40, '/static/cards/design-09/IMG_4454.jpg', 'wedding-cards'),
('Floral Vellum Transparency', 'Translucent vellum invitation with delicate wildflower watercolor print, pearl accents, wax seal — ethereal and romantic', 250, 40, '/static/cards/design-10/IMG_4536.jpg', 'wedding-cards'),
('Mughal Garden Invitation', 'Textured cream envelope with gold botanical toile print, Mughal palace illustration liner, orange jali pattern — traditional South Asian grandeur', 250, 40, '/static/cards/design-11/IMG_8080.jpg', 'wedding-cards'),
('Burgundy Arch Shalima', 'Cream arch-shaped shalima ceremony invitation with burgundy border, elegant calligraphy, QR code, and botanical leaf accents', 250, 40, '/static/cards/design-12/homepage3.jpg', 'wedding-cards'),
('Mughal Rose Wax Seal Suite', 'Embossed cream card with Mughal arch design, rose garden and minaret illustrations, vellum wrap with gold wax seal and baby''s breath', 250, 40, '/static/cards/design-13/IMG_9586.jpg', 'wedding-cards'),
('Forest Green Oval Walima', 'Dark forest green oval-shaped cards with cream embossed walima invitation, gold leaf accent, crest monogram, QR code — modern meets classic', 250, 40, '/static/cards/design-14/IMG_9635.jpg', 'wedding-cards');

-- Seed card_images for each design (gallery photos)
-- design-01
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-01/27DEBC49-9228-41D7-B621-873B77C8E73F.JPG', 1 FROM cards WHERE name = 'Vibrant Dholki Booklet';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-01/34AD430F-63AA-405E-B0B6-842FA01A0854.JPG', 2 FROM cards WHERE name = 'Vibrant Dholki Booklet';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-01/CB557ED6-92CC-4503-BD90-DA91E4C24E28.JPG', 3 FROM cards WHERE name = 'Vibrant Dholki Booklet';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-01/EEAEDE0E-A6D9-45B5-A27A-31A20098F7E4.JPG', 4 FROM cards WHERE name = 'Vibrant Dholki Booklet';

-- design-02
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-02/homepage1.JPG', 1 FROM cards WHERE name = 'Sage & Gold Nikkah Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-02/74041FB5-C964-4E0F-8AF0-BD49B1899929.JPG', 2 FROM cards WHERE name = 'Sage & Gold Nikkah Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-02/802902AC-E832-4E28-8033-00BF3CEB179A.JPG', 3 FROM cards WHERE name = 'Sage & Gold Nikkah Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-02/85A2CF55-3566-443F-8058-07BB2A447894.JPG', 4 FROM cards WHERE name = 'Sage & Gold Nikkah Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-02/967359F5-7EE3-452F-9F6E-B40292DB10D8.JPG', 5 FROM cards WHERE name = 'Sage & Gold Nikkah Suite';

-- design-03
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-03/homepage2.jpg', 1 FROM cards WHERE name = 'Blush Mehendi Laser-Cut';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-03/IMG_0271.jpg', 2 FROM cards WHERE name = 'Blush Mehendi Laser-Cut';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-03/IMG_0275.jpg', 3 FROM cards WHERE name = 'Blush Mehendi Laser-Cut';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-03/IMG_0291.jpg', 4 FROM cards WHERE name = 'Blush Mehendi Laser-Cut';

-- design-04
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-04/IMG_0304.jpg', 1 FROM cards WHERE name = 'Blush Mehendi Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-04/IMG_0327.jpg', 2 FROM cards WHERE name = 'Blush Mehendi Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-04/IMG_0330.jpg', 3 FROM cards WHERE name = 'Blush Mehendi Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-04/IMG_0356.jpg', 4 FROM cards WHERE name = 'Blush Mehendi Wax Seal Suite';

-- design-05
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-05/IMG_0449.jpg', 1 FROM cards WHERE name = 'Sage Monogram Envelope';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-05/IMG_0452.jpg', 2 FROM cards WHERE name = 'Sage Monogram Envelope';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-05/IMG_0468.jpg', 3 FROM cards WHERE name = 'Sage Monogram Envelope';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-05/IMG_0471.jpg', 4 FROM cards WHERE name = 'Sage Monogram Envelope';

-- design-06
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-06/IMG_1679.jpg', 1 FROM cards WHERE name = 'Emerald Nikkah Dua Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-06/IMG_1686.jpg', 2 FROM cards WHERE name = 'Emerald Nikkah Dua Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-06/IMG_1715.jpg', 3 FROM cards WHERE name = 'Emerald Nikkah Dua Suite';

-- design-07
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-07/IMG_1721.jpg', 1 FROM cards WHERE name = 'Royal Blue Silver Crest';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-07/IMG_1725.jpg', 2 FROM cards WHERE name = 'Royal Blue Silver Crest';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-07/IMG_1746.jpg', 3 FROM cards WHERE name = 'Royal Blue Silver Crest';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-07/IMG_1752.jpg', 4 FROM cards WHERE name = 'Royal Blue Silver Crest';

-- design-08
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-08/IMG_4429.jpg', 1 FROM cards WHERE name = 'Emerald Baraat Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-08/IMG_4431.jpg', 2 FROM cards WHERE name = 'Emerald Baraat Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-08/IMG_4433.jpg', 3 FROM cards WHERE name = 'Emerald Baraat Suite';

-- design-09
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-09/IMG_4454.jpg', 1 FROM cards WHERE name = 'Burgundy Shalima Elegance';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-09/IMG_4505.JPG.jpeg', 2 FROM cards WHERE name = 'Burgundy Shalima Elegance';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-09/IMG_4506.JPG.jpeg', 3 FROM cards WHERE name = 'Burgundy Shalima Elegance';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-09/IMG_4509.JPG.jpeg', 4 FROM cards WHERE name = 'Burgundy Shalima Elegance';

-- design-10
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-10/IMG_4536.jpg', 1 FROM cards WHERE name = 'Floral Vellum Transparency';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-10/IMG_4541.jpg', 2 FROM cards WHERE name = 'Floral Vellum Transparency';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-10/IMG_4544.jpg', 3 FROM cards WHERE name = 'Floral Vellum Transparency';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-10/IMG_4545.jpg', 4 FROM cards WHERE name = 'Floral Vellum Transparency';

-- design-11
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-11/IMG_8080.jpg', 1 FROM cards WHERE name = 'Mughal Garden Invitation';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-11/IMG_8084.jpg', 2 FROM cards WHERE name = 'Mughal Garden Invitation';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-11/IMG_8134.jpg', 3 FROM cards WHERE name = 'Mughal Garden Invitation';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-11/IMG_8152.jpg', 4 FROM cards WHERE name = 'Mughal Garden Invitation';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-11/IMG_8153.jpg', 5 FROM cards WHERE name = 'Mughal Garden Invitation';

-- design-12
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-12/homepage3.jpg', 1 FROM cards WHERE name = 'Burgundy Arch Shalima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-12/IMG_9480.jpg', 2 FROM cards WHERE name = 'Burgundy Arch Shalima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-12/IMG_9493.jpg', 3 FROM cards WHERE name = 'Burgundy Arch Shalima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-12/IMG_9496.jpg', 4 FROM cards WHERE name = 'Burgundy Arch Shalima';

-- design-13
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-13/IMG_9586.jpg', 1 FROM cards WHERE name = 'Mughal Rose Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-13/IMG_9589.jpg', 2 FROM cards WHERE name = 'Mughal Rose Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-13/IMG_9610.jpg', 3 FROM cards WHERE name = 'Mughal Rose Wax Seal Suite';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-13/IMG_9630.jpg', 4 FROM cards WHERE name = 'Mughal Rose Wax Seal Suite';

-- design-14
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-14/IMG_9635.jpg', 1 FROM cards WHERE name = 'Forest Green Oval Walima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-14/IMG_9636.jpg', 2 FROM cards WHERE name = 'Forest Green Oval Walima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-14/IMG_9638.jpg', 3 FROM cards WHERE name = 'Forest Green Oval Walima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-14/IMG_9804.jpg', 4 FROM cards WHERE name = 'Forest Green Oval Walima';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/cards/design-14/IMG_9813.jpg', 5 FROM cards WHERE name = 'Forest Green Oval Walima';

-- migrate:down
DELETE FROM card_images;
DROP TABLE card_images;
DROP INDEX idx_cards_category;
ALTER TABLE cards DROP COLUMN category;
ALTER TABLE cards DROP COLUMN price_pkr;
ALTER TABLE cards DROP COLUMN price_nok;
ALTER TABLE cards ADD COLUMN price BIGINT NOT NULL DEFAULT 0;
ALTER TABLE orders DROP COLUMN currency;
