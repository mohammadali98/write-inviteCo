-- name: GetOrderByID :one
SELECT
    o.id,
    o.customer_id,
    o.card_id,
    o.quantity,
    o.total_price,
    o.status,
    o.currency,
    o.created_at,
    o.updated_at,
    COALESCE(c.name, '') AS card_name,
    COALESCE(c.image, '') AS card_image
FROM orders o
LEFT JOIN cards c ON o.card_id = c.id
WHERE o.id = $1;

-- name: GetOrdersByCustomerID :many
SELECT id, customer_id, card_id, quantity, total_price, status, currency, created_at, updated_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: GetAdminOrders :many
SELECT
    o.id,
    COALESCE(c.name, 'Unknown Customer') AS customer_name,
    o.total_price,
    o.status,
    o.currency,
    o.created_at
FROM orders o
LEFT JOIN customers c ON c.id = o.customer_id
ORDER BY o.created_at DESC;

-- name: GetLatestOrderDetailByOrderID :many
SELECT
    id,
    order_id,
    COALESCE(side, 'bride') AS side,
    bride_name,
    groom_name,
    bride_father_name,
    groom_father_name,
    COALESCE(mehndi_date::text, '')::text AS mehndi_date,
    COALESCE(mehndi_day, '') AS mehndi_day,
    mehndi_time_type,
    COALESCE(mehndi_time::text, '')::text AS mehndi_time,
    COALESCE(mehndi_dinner_time::text, '')::text AS mehndi_dinner_time,
    COALESCE(mehndi_venue_name, '') AS mehndi_venue_name,
    COALESCE(mehndi_venue_address, '') AS mehndi_venue_address,
    COALESCE(baraat_date::text, '')::text AS baraat_date,
    COALESCE(baraat_day, '') AS baraat_day,
    baraat_time_type,
    COALESCE(baraat_time::text, '')::text AS baraat_time,
    COALESCE(baraat_dinner_time::text, '')::text AS baraat_dinner_time,
    COALESCE(baraat_arrival_time::text, '')::text AS baraat_arrival_time,
    COALESCE(rukhsati_time::text, '')::text AS rukhsati_time,
    COALESCE(baraat_venue_name, '') AS baraat_venue_name,
    COALESCE(baraat_venue_address, '') AS baraat_venue_address,
    COALESCE(nikkah_date::text, '')::text AS nikkah_date,
    COALESCE(nikkah_day, '') AS nikkah_day,
    nikkah_time_type,
    COALESCE(nikkah_time::text, '')::text AS nikkah_time,
    COALESCE(nikkah_dinner_time::text, '')::text AS nikkah_dinner_time,
    COALESCE(nikkah_venue_name, '') AS nikkah_venue_name,
    COALESCE(nikkah_venue_address, '') AS nikkah_venue_address,
    COALESCE(walima_date::text, '')::text AS walima_date,
    COALESCE(walima_day, '') AS walima_day,
    walima_time_type,
    COALESCE(walima_time::text, '')::text AS walima_time,
    COALESCE(walima_dinner_time::text, '')::text AS walima_dinner_time,
    COALESCE(reception_time::text, '')::text AS reception_time,
    COALESCE(walima_venue_name, '') AS walima_venue_name,
    COALESCE(walima_venue_address, '') AS walima_venue_address,
    COALESCE(rsvp_name, '') AS rsvp_name,
    COALESCE(rsvp_phone, '') AS rsvp_phone,
    notes,
    created_at
FROM order_details
WHERE order_id = $1
ORDER BY created_at DESC
LIMIT 1;
