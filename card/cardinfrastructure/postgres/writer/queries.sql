-- name: CreateCard :one
INSERT INTO cards (name, description, price_foil_pkr, price_nofoil_pkr, price_foil_nok, price_nofoil_nok, insert_price_pkr, insert_price_nok, min_order, included_inserts, image, category)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, name, description, price_foil_pkr, price_nofoil_pkr, price_foil_nok, price_nofoil_nok, insert_price_pkr, insert_price_nok, min_order, included_inserts, image, category, created_at, updated_at;

-- name: UpdateCard :exec
UPDATE cards
SET name = $2,
    description = $3,
    price_foil_pkr = $4,
    price_nofoil_pkr = $5,
    price_foil_nok = $6,
    price_nofoil_nok = $7,
    insert_price_pkr = $8,
    insert_price_nok = $9,
    min_order = $10,
    included_inserts = $11,
    image = $12,
    category = $13,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteCard :exec
DELETE FROM cards
WHERE id = $1;

-- name: CreateCardImage :one
INSERT INTO card_images (card_id, image, sort_order)
VALUES ($1, $2, $3)
RETURNING id, card_id, image, sort_order, created_at;

-- name: DeleteCardImagesByCardID :exec
DELETE FROM card_images
WHERE card_id = $1;
