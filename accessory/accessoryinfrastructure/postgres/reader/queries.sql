-- name: ListActiveAccessories :many
SELECT id, name, category, description, is_active, created_at, updated_at
FROM accessories
WHERE is_active = true
ORDER BY created_at DESC;

-- name: GetAccessoryImagesByAccessoryID :many
SELECT id, accessory_id, image_url, sort_order, created_at
FROM accessory_images
WHERE accessory_id = $1
ORDER BY sort_order ASC;
