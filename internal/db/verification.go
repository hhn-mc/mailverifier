package db

import (
	"time"

	"github.com/hhn-mc/mailverifier/internal/player"
	"golang.org/x/net/context"
)

func (db *DB) LatestVerification(playerUUID string) (player.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	row := db.QueryRow(ctx, `
SELECT id, player_uuid, created_at
FROM verifications
WHERE player_uuid = $1
ORDER BY created_at DESC
LIMIT 1
`, playerUUID)

	var v player.Verification
	if err := row.Scan(&v.ID, &v.PlayerUUID, &v.CreatedAt); err != nil {
		return player.Verification{}, err
	}

	emails, err := db.VerificationEmails(v.ID)
	if err != nil {
		return player.Verification{}, err
	}
	v.Emails = emails

	for _, email := range emails {
		if email.VerifiedAt != nil {
			v.IsVerified = true
			break
		}
	}

	return v, nil
}

func (db *DB) HasVerification(playerUUID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	res, err := db.Exec(ctx, `
SELECT created_at
FROM verifications
WHERE player_uuid = $1;
`, playerUUID)

	return res.RowsAffected() > 0, err
}

func (db *DB) Verifications(playerUUID string) ([]player.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	rows, err := db.Query(ctx, `
SELECT id, player_uuid, created_at
FROM verifications
WHERE player_uuid = $1
`, playerUUID)
	if err != nil {
		return nil, err
	}

	var vv []player.Verification
	for rows.Next() {
		var v player.Verification
		if err := rows.Scan(&v.ID, &v.PlayerUUID, &v.CreatedAt); err != nil {
			return nil, err
		}

		emails, err := db.VerificationEmails(v.ID)
		if err != nil {
			return nil, err
		}
		v.Emails = emails

		for _, email := range emails {
			if email.VerifiedAt != nil {
				v.IsVerified = true
				break
			}
		}

		vv = append(vv, v)
	}
	return vv, nil
}

func (db *DB) VerificationEmails(verificationID uint64) ([]player.VerificationEmail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	rows, err := db.Query(ctx, `
SELECT email, verified_at, created_at
FROM verification_emails
WHERE verification_id = $1
`, verificationID)
	if err != nil {
		return nil, err
	}

	var ee []player.VerificationEmail
	for rows.Next() {
		var e player.VerificationEmail
		verifiedAt := &time.Time{}
		if err := rows.Scan(&e.Email, &verifiedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		if verifiedAt != nil {
			e.VerifiedAt = verifiedAt
		}
		ee = append(ee, e)
	}
	return ee, nil
}

func (db *DB) CreateVerification(v *player.Verification) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	return db.QueryRow(ctx, `
INSERT INTO verifications
(player_uuid)
VALUES ($1)
RETURNING id, created_at;
`, v.PlayerUUID).
		Scan(&v.ID, &v.CreatedAt)
}

func (db *DB) CreateEmailVerification(verificationID uint64, code, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	_, err := db.Exec(ctx, `
INSERT INTO verification_emails
(verification_id, code, email)
VALUES ($1, $2, $3);
`, verificationID, code, email)
	return err
}
