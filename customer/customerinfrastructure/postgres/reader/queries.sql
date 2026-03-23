-- name: GetCustomerByID :one
SELECT id, name, email, phone, created_at, updated_at
FROM customers
WHERE id = $1;
