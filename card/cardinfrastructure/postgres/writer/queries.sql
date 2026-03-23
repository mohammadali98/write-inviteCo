-- name: CreateCard :one
INSERT INTO cards (name, description, price, image)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, price, image, created_at, updated_at;

-- name: UpdateCard :exec
UPDATE cards
SET name = $2, description = $3, price = $4, image = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteCard :exec
DELETE FROM cards
WHERE id = $1;
