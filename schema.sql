CREATE TABLE IF NOT EXISTS players
(
    uuid UUID NOT NULL PRIMARY KEY,
    username VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verifications
(
    token UUID NOT NULL PRIMARY KEY,
    player_uuid UUID REFERENCES players (uuid) NOT NULL,
    email_token UUID,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verification_emails
(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    verification_token UUID REFERENCES verifications (token) NOT NULL,
    email VARCHAR(320) NOT NULL,
);