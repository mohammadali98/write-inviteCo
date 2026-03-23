-- migrate:up
CREATE TABLE cards
(
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT,
    price       BIGINT      NOT NULL,
    image       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- migrate:down
DROP TABLE cards;
