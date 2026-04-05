-- migrate:up
CREATE TABLE order_details
(
    id                BIGSERIAL PRIMARY KEY,
    order_id          BIGINT      NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    bride_name        TEXT        NOT NULL,
    groom_name        TEXT        NOT NULL,
    bride_father_name TEXT        NOT NULL,
    groom_father_name TEXT        NOT NULL,
    event_type        TEXT        NOT NULL,
    event_date        TEXT        NOT NULL,
    event_time        TEXT        NOT NULL,
    venue_name        TEXT        NOT NULL,
    venue_address     TEXT        NOT NULL,
    rsvp_name         TEXT        NOT NULL,
    rsvp_phone        TEXT        NOT NULL,
    notes             TEXT,
    created_at        TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_order_details_order_id ON order_details (order_id);

-- migrate:down
DROP INDEX idx_order_details_order_id;
DROP TABLE order_details;
