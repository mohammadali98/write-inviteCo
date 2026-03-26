-- name: GetCustomerByID :one
SELECT id, name, email, phone, address, city, postal_code, created_at, updated_at
FROM customers
WHERE id = $1;
