package main

import (
	"errors"
	"time"

	_ "embed"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

//go:embed schema.sql
var schema string

type database struct {
	dsn     string
	timeout time.Duration

	*pgxpool.Pool
}

func (db *database) open() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var err error
	cfg, err := pgxpool.ParseConfig(db.dsn)
	if err != nil {
		return err
	}

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	db.Pool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}

	return db.Ping(ctx)
}

func (db *database) migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	_, err := db.Exec(ctx, schema)
	if err != nil {
		return err
	}
	return nil
}

func (db *database) getActiveVerificationIDAndCodeForPlayerUUID(playerUUID string) (uint64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var id uint64
	var code string
	if err := db.QueryRow(ctx, `
SELECT id, code
FROM verifications
WHERE player_uuid = $1
AND verified_at IS NULL;`, playerUUID).Scan(&id, &code); err != nil {
		return 0, "", err
	}
	return id, code, nil
}

func (db *database) getPlayerUUIDFromToken(token string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var playerUUID pgtypeuuid.UUID
	if err := db.QueryRow(ctx, `
SELECT player_uuid
FROM verifications
WHERE v.token = $1;`, token).Scan(&playerUUID); err != nil {
		return uuid.Nil, err
	}
	return playerUUID.UUID, nil
}

func (db *database) getPlayerUsernameFromToken(token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var username string
	if err := db.QueryRow(ctx, `
SELECT p.username
FROM players p, verifications v
WHERE v.token = $1
AND v.player_uuid = p.uuid;`, token).Scan(&username); err != nil {
		return "", err
	}
	return username, nil
}

func (db *database) createPlayer(uuid, username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	_, err := db.Exec(ctx, `
INSERT INTO players
(uuid, username)
VALUES ($1, $2);`, uuid, username)
	return err
}

func (db *database) doesPlayerExist(uuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var createdAt time.Time
	if err := db.QueryRow(ctx, `
SELECT created_at
FROM players
WHERE uuid = $1;`, uuid).Scan(&createdAt); err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *database) createVerification(playerUUID, code string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var id uint64
	err := db.QueryRow(ctx, `
INSERT INTO verifications
(player_uuid, code)
VALUES ($1, $2)
RETURNING id;`, playerUUID, code).Scan(&id)
	return id, err
}

func (db *database) createEmailVerification(verificationID uint64, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	if _, err := db.Exec(ctx, `
INSERT INTO verification_emails
(verification_id, email)
VALUES ($1, $2);`, verificationID, email); err != nil {
		return err
	}
	return nil
}

func (db *database) doesVerificationTokenExist(token, playerUUID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	var createdAt time.Time
	if err := db.QueryRow(ctx, `
SELECT created_at
FROM verifications
WHERE token = $1
AND player_uuid = $2
AND verified_at IS NULL`,
		token, playerUUID).Scan(&createdAt); err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *database) insertEmailVerificationToken(token, emailToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	if _, err := db.Exec(ctx, `
INSERT INTO verifications
(email_token)
VALUES ($1)
WHERE token = $2;`, emailToken, token); err != nil {
		return err
	}
	return nil
}

func (db *database) insertVerifiedEmail(token, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	if _, err := db.Exec(ctx, `
INSERT INTO verifications
(email, verified_at)
VALUES ($1, CURRENT_TIMESTAMP)
WHERE token = $2;`, email, token); err != nil {
		return err
	}
	return nil
}
