-- name: CreateCustomer :one
INSERT INTO customers (name, email, phone)
VALUES ($1, $2, $3)
RETURNING id, name, email, phone, created_at, updated_at;
