-- name: ListProducts :many
SELECT id, card_id, name, category, price, image_url, description, dimensions, is_active
FROM products
ORDER BY id DESC;

-- name: GetProduct :one
SELECT id, card_id, name, category, price, image_url, description, dimensions, is_active
FROM products
WHERE id = $1;

-- name: GetProductImages :many
SELECT image_url
FROM product_images
WHERE product_id = $1
ORDER BY sort_order ASC, id ASC;
