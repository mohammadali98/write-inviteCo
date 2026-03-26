-- name: GetAllCards :many
SELECT id, name, description, price_pkr, price_nok, image, category, created_at, updated_at
FROM cards
ORDER BY created_at DESC;

-- name: GetCardByID :one
SELECT id, name, description, price_pkr, price_nok, image, category, created_at, updated_at
FROM cards
WHERE id = $1;

-- name: GetCardsByCategory :many
SELECT id, name, description, price_pkr, price_nok, image, category, created_at, updated_at
FROM cards
WHERE category = $1
ORDER BY created_at DESC;

-- name: SearchCards :many
SELECT id, name, description, price_pkr, price_nok, image, category, created_at, updated_at
FROM cards
WHERE name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%'
ORDER BY created_at DESC;

-- name: GetCardImagesByCardID :many
SELECT id, card_id, image, sort_order, created_at
FROM card_images
WHERE card_id = $1
ORDER BY sort_order ASC;
