CREATE TABLE IF NOT EXISTS players
(
    uuid UUID NOT NULL PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verifications
(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    player_uuid UUID REFERENCES players (uuid) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verification_emails
(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    verification_id BIGINT REFERENCES verifications (id),
    code TEXT NOT NULL,
    email TEXT NOT NULL,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (verification_id, code, email)
);
