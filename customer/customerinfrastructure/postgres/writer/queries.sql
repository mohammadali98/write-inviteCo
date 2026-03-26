-- name: CreateCustomer :one
INSERT INTO customers (name, email, phone, address, city, postal_code)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, email, phone, address, city, postal_code, created_at, updated_at;
