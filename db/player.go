package db

import (
	"errors"
	"time"

	"github.com/hhn-mc/mailverifier/player"
	"github.com/jackc/pgx/v4"
	"golang.org/x/net/context"
)

func (db *DB) PlayerWithUUIDExists(uuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	res, err := db.Exec(ctx, `
SELECT created_at
FROM players
WHERE uuid = $1;
`, uuid)

	return res.RowsAffected() > 0, err
}

func (db *DB) PlayerByUUID(uuid string) (player.Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	row := db.QueryRow(ctx, `
SELECT uuid, username, created_at
FROM players
WHERE uuid = $1
`, uuid)

	var p player.Player
	if err := row.Scan(&p.UUID, &p.Username, &p.CreatedAt); err != nil {
		return player.Player{}, err
	}

	row = db.QueryRow(ctx, `
SELECT created_at
FROM verification_emails
WHERE verification_id = (
	SELECT id
	FROM verifications
	WHERE player_uuid = $1
	ORDER BY created_at DESC
	LIMIT 1
) AND verified_at IS NOT NULL
`, uuid)

	p.IsVerified = true
	if err := row.Scan(&time.Time{}); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return player.Player{}, err
		}
		p.IsVerified = false
	}

	return p, nil
}

func (db *DB) CreatePlayer(p *player.Player) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	return db.QueryRow(ctx, `
INSERT INTO players
(uuid, username)
VALUES ($1, $2)
RETURNING created_at;
`, p.UUID, p.Username).
		Scan(&p.CreatedAt)
}
