-- name: GetOrderByID :one
SELECT
    o.id,
    o.customer_id,
    o.card_id,
    o.quantity,
    o.total_price,
    o.status,
    o.currency,
    o.public_token::text AS public_token,
    o.created_at,
    o.updated_at,
    COALESCE(c.name, '') AS card_name,
    COALESCE(c.image, '') AS card_image,
    COALESCE(c.category, '') AS card_category
FROM orders o
LEFT JOIN cards c ON o.card_id = c.id
WHERE o.id = $1;

-- name: GetOrderByPublicToken :one
SELECT
    o.id,
    o.customer_id,
    o.card_id,
    o.quantity,
    o.total_price,
    o.status,
    o.currency,
    o.public_token::text AS public_token,
    o.created_at,
    o.updated_at,
    COALESCE(c.name, '') AS card_name,
    COALESCE(c.image, '') AS card_image,
    COALESCE(c.category, '') AS card_category
FROM orders o
LEFT JOIN cards c ON o.card_id = c.id
WHERE o.public_token::text = sqlc.arg(public_token)::text;

-- name: GetOrdersByCustomerID :many
SELECT id, customer_id, card_id, quantity, total_price, status, currency, created_at, updated_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: GetAdminOrders :many
SELECT
    o.id,
    COALESCE(cu.name, 'Unknown Customer') AS customer_name,
    COALESCE(ca.name, '') AS product_name,
    COALESCE(ca.category, '') AS card_category,
    o.quantity,
    o.total_price,
    o.status,
    op.payment_status,
    op.submitted_amount,
    op.submitted_at,
    o.currency,
    o.created_at
FROM orders o
LEFT JOIN customers cu ON cu.id = o.customer_id
LEFT JOIN cards ca ON ca.id = o.card_id
LEFT JOIN order_payments op ON op.order_id = o.id
WHERE
    (sqlc.narg(order_status)::text IS NULL OR o.status = sqlc.narg(order_status)::text)
    AND (sqlc.narg(payment_status)::text IS NULL OR op.payment_status = sqlc.narg(payment_status)::text)
    AND (
        sqlc.narg(search)::text IS NULL
        OR cu.name ILIKE '%' || sqlc.narg(search)::text || '%'
        OR cu.phone ILIKE '%' || sqlc.narg(search)::text || '%'
        OR o.id::text ILIKE '%' || sqlc.narg(search)::text || '%'
    )
    AND (sqlc.narg(created_from)::timestamptz IS NULL OR o.created_at >= sqlc.narg(created_from)::timestamptz)
    AND (sqlc.narg(created_to)::timestamptz IS NULL OR o.created_at < sqlc.narg(created_to)::timestamptz)
ORDER BY o.created_at DESC;

-- name: GetLatestOrderDetailByOrderID :many
SELECT
    id,
    order_id,
    COALESCE(side, 'bride') AS side,
    COALESCE(extra_inserts_per_card, 0)::bigint AS extra_inserts_per_card,
    COALESCE(top_label, '') AS top_label,
    COALESCE(couple_name, '') AS couple_name,
    COALESCE(event_date, '') AS event_date,
    COALESCE(bid_box_details, '') AS bid_box_details,
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
    COALESCE(baraat_sehrabandi_time::text, '')::text AS baraat_sehrabandi_time,
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
    COALESCE(shendi_date::text, '')::text AS shendi_date,
    COALESCE(shendi_day, '') AS shendi_day,
    COALESCE(shendi_time::text, '')::text AS shendi_time,
    shendi_time_type,
    COALESCE(shendi_dinner_time::text, '')::text AS shendi_dinner_time,
    COALESCE(shendi_venue_name, '') AS shendi_venue_name,
    COALESCE(shendi_venue_address, '') AS shendi_venue_address,
    COALESCE(shendi_arrival_time::text, '')::text AS shendi_arrival_time,
    COALESCE(shendi_rukhsati_time::text, '')::text AS shendi_rukhsati_time,
    COALESCE(shendi_sehrabandi_time::text, '')::text AS shendi_sehrabandi_time,
    COALESCE(shalima_date::text, '')::text AS shalima_date,
    COALESCE(shalima_day, '') AS shalima_day,
    COALESCE(shalima_time::text, '')::text AS shalima_time,
    shalima_time_type,
    COALESCE(shalima_dinner_time::text, '')::text AS shalima_dinner_time,
    COALESCE(shalima_venue_name, '') AS shalima_venue_name,
    COALESCE(shalima_venue_address, '') AS shalima_venue_address,
    COALESCE(shalima_arrival_time::text, '')::text AS shalima_arrival_time,
    COALESCE(shalima_rukhsati_time::text, '')::text AS shalima_rukhsati_time,
    COALESCE(shalima_sehrabandi_time::text, '')::text AS shalima_sehrabandi_time,
    COALESCE(shalima_reception_time::text, '')::text AS shalima_reception_time,
    COALESCE(rsvp_name, '') AS rsvp_name,
    COALESCE(rsvp_phone, '') AS rsvp_phone,
    notes,
    created_at
FROM order_details
WHERE order_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetOrderPaymentByOrderID :one
SELECT
    id,
    order_id,
    payment_method,
    payment_status,
    expected_amount,
    submitted_amount,
    sender_name,
    transaction_reference,
    proof_file_path,
    customer_note,
    submitted_at,
    verified_at,
    rejected_at,
    admin_note,
    created_at,
    updated_at
FROM order_payments
WHERE order_id = $1;
