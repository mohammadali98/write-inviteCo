-- migrate:up
CREATE TABLE accessories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'wax-seal',
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accessory_images (
    id BIGSERIAL PRIMARY KEY,
    accessory_id BIGINT NOT NULL REFERENCES accessories(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accessory_images_accessory_id ON accessory_images (accessory_id);

-- migrate:down
DROP TABLE accessory_images;
DROP TABLE accessories;
