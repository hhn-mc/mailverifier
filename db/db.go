package db

import (
	"errors"
	"time"

	_ "embed"

	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

//go:embed schema.sql
var schema string

type DB struct {
	dsn     string
	timeout time.Duration

	*pgxpool.Pool
}

func (db *DB) Open() error {
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

	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	db.Pool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}

	return db.Ping(ctx)
}

func (db *DB) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	_, err := db.Exec(ctx, schema)
	return err
}

func (db *DB) getUnverifiedValidationID(playerUUID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	row := db.QueryRow(ctx, `
SELECT id
FROM verifications
WHERE player_uuid = $1
AND verified_at IS NULL;`,
		playerUUID)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *DB) createValidation(playerUUID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	row := db.QueryRow(ctx, `
INSERT INTO verifications
(uuid)
VALUES ($1)
RETURNING id;`,
		playerUUID)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *DB) getVerificationEmailID(validationID uint64, code string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	row := db.QueryRow(ctx, `
SELECT id
FROM verification_emails
WHERE verification_id = $1
AND code = $2;`,
		validationID, code)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *DB) createEmailVerification(verificationID uint64, code, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	_, err := db.Exec(ctx, `
INSERT INTO verification_emails
(verification_id, code, email)
VALUES ($1, $2, $3);`,
		verificationID, code, email)
	return err
}

func (db *DB) doesVerificationTokenExist(token, playerUUID string) (bool, error) {
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

func (db *DB) insertEmailVerificationToken(token, emailToken string) error {
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

func (db *DB) insertVerifiedEmail(token, email string) error {
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
