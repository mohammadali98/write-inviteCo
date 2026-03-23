-- name: GetAllCards :many
SELECT id, name, description, price, image, created_at, updated_at
FROM cards
ORDER BY created_at DESC;

-- name: GetCardByID :one
SELECT id, name, description, price, image, created_at, updated_at
FROM cards
WHERE id = $1;
