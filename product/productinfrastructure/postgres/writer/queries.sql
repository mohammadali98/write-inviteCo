-- name: CreateProduct :one
INSERT INTO products (card_id, name, category, price, image_url, description, dimensions, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, card_id, name, category, price, image_url, description, dimensions, is_active;

-- name: UpdateProduct :exec
UPDATE products
SET card_id = $2,
    name = $3,
    category = $4,
    price = $5,
    image_url = $6,
    description = $7,
    dimensions = $8,
    is_active = $9,
    updated_at = NOW()
WHERE id = $1;

-- name: AddProductImage :exec
INSERT INTO product_images (product_id, image_url, sort_order)
VALUES ($1, $2, $3);

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;
