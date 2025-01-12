-- +migrate Up

CREATE TABLE IF NOT EXISTS decisions (
    id SERIAL NOT NULL,
    recipient_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    liked BOOLEAN NOT NULL,
    mutually_liked BOOLEAN,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    UNIQUE(recipient_id, actor_id)
);

-- +migrate Down

DROP TABLE IF EXISTS decisions;

