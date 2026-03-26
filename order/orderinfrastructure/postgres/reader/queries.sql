-- name: GetOrderByID :one
SELECT id, customer_id, card_id, quantity, total_price, status, currency, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrdersByCustomerID :many
SELECT id, customer_id, card_id, quantity, total_price, status, currency, created_at, updated_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;
