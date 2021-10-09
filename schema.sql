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
    verification_emails_id BIGINT REFERENCES verification_emails (id),
    code TEXT NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verification_emails
(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    verification_id BIGINT REFERENCES verifications (id) NOT NULL,
    email TEXT NOT NULL
);