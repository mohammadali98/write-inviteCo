-- name: CreateOrder :one
INSERT INTO orders (customer_id, card_id, quantity, total_price, status, currency)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, customer_id, card_id, quantity, total_price, status, currency, created_at, updated_at;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
