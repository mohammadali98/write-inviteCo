-- migrate:up

-- Seed bid box and potli products
INSERT INTO cards (name, description, price_pkr, price_nok, image, category) VALUES
('Ivory Floral Acrylic Bid Box', 'Clear acrylic bid boxes styled with ivory ribbons, floral accents, and custom name tags for elegant wedding gifting.', 450, 70, '/static/bid-boxes/78D64C8C-F06F-4FE9-9CEF-6FA2E79FAEED.JPG', 'bid-boxes'),
('Ruby Gold Acrylic Bid Box', 'Transparent favor boxes with ruby and gold ribbons, white florals, and personalized tags for vibrant mehendi and dholki events.', 450, 70, '/static/bid-boxes/f186604b-c7cb-42a0-a58d-c90055d28f40.JPG', 'bid-boxes'),
('Rose Quartz Ribbon Bid Box', 'Rose quartz themed acrylic bid boxes finished with satin bows and pearl accents for premium guest favors.', 450, 70, '/static/bid-boxes/IMG_0044.jpg', 'bid-boxes'),
('Peach Festive Potli Set', 'Handcrafted peach potli favor pouches with gold thread work and sequined detailing, ideal for festive wedding giveaways.', 350, 55, '/static/bid-boxes/IMG_1978.jpg', 'bid-boxes'),
('Sunset Organza Potli Bundle', 'Soft organza potli bags with shimmer lace edging and braided drawstrings in warm sunset tones.', 350, 55, '/static/bid-boxes/IMG_1790.jpg', 'bid-boxes'),
('Lavender Floral Favor Tins', 'Round floral favor tins with custom top labels and gold rims for sweets, dry fruits, or keepsakes.', 300, 50, '/static/bid-boxes/IMG_9309.jpg', 'bid-boxes');

-- card_images for bid box and potli product galleries
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/78D64C8C-F06F-4FE9-9CEF-6FA2E79FAEED.JPG', 1 FROM cards WHERE name = 'Ivory Floral Acrylic Bid Box';

INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/f186604b-c7cb-42a0-a58d-c90055d28f40.JPG', 1 FROM cards WHERE name = 'Ruby Gold Acrylic Bid Box';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_0125.jpg', 2 FROM cards WHERE name = 'Ruby Gold Acrylic Bid Box';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_0132.jpg', 3 FROM cards WHERE name = 'Ruby Gold Acrylic Bid Box';

INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_0044.jpg', 1 FROM cards WHERE name = 'Rose Quartz Ribbon Bid Box';

INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_1978.jpg', 1 FROM cards WHERE name = 'Peach Festive Potli Set';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_1985.jpg', 2 FROM cards WHERE name = 'Peach Festive Potli Set';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_1997.jpg', 3 FROM cards WHERE name = 'Peach Festive Potli Set';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_1998.jpg', 4 FROM cards WHERE name = 'Peach Festive Potli Set';

INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_1790.jpg', 1 FROM cards WHERE name = 'Sunset Organza Potli Bundle';

INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_9309.jpg', 1 FROM cards WHERE name = 'Lavender Floral Favor Tins';
INSERT INTO card_images (card_id, image, sort_order) SELECT id, '/static/bid-boxes/IMG_9320.jpg', 2 FROM cards WHERE name = 'Lavender Floral Favor Tins';

-- migrate:down
DELETE FROM card_images
WHERE card_id IN (
    SELECT id FROM cards WHERE name IN (
        'Ivory Floral Acrylic Bid Box',
        'Ruby Gold Acrylic Bid Box',
        'Rose Quartz Ribbon Bid Box',
        'Peach Festive Potli Set',
        'Sunset Organza Potli Bundle',
        'Lavender Floral Favor Tins'
    )
);

DELETE FROM cards
WHERE name IN (
    'Ivory Floral Acrylic Bid Box',
    'Ruby Gold Acrylic Bid Box',
    'Rose Quartz Ribbon Bid Box',
    'Peach Festive Potli Set',
    'Sunset Organza Potli Bundle',
    'Lavender Floral Favor Tins'
);
