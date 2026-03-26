-- name: CreateCard :one
INSERT INTO cards (name, description, price_pkr, price_nok, image, category)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, description, price_pkr, price_nok, image, category, created_at, updated_at;

-- name: UpdateCard :exec
UPDATE cards
SET name = $2, description = $3, price_pkr = $4, price_nok = $5, image = $6, category = $7, updated_at = CURRENT_TIMESTAMP
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
