package db

import (
	"github.com/hhn-mc/mailverifier/player"
	"golang.org/x/net/context"
)

func (db *DB) PlayerWithUUIDExists(uuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	res, err := db.Exec(ctx, `
SELECT uuid, username, created_at
FROM players
WHERE uuid = $1;
`, uuid)

	return res.RowsAffected() > 0, err
}

func (db *DB) PlayerByUUID(uuid string) (player.Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	var p player.Player
	return p, db.QueryRow(ctx, `
SELECT uuid, 
	(CASE 
		WHEN verification_id IS NULL THEN TRUE
		ELSE FALSE
	END),
	username,
	created_at
FROM players
WHERE uuid = $1;
`, uuid).
		Scan(&p.UUID, &p.IsVerified, &p.Username, &p.CreatedAt)
}

func (db *DB) CreatePlayer(p *player.Player) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()

	return db.QueryRow(ctx, `
INSERT INTO players
(uuid, username)
VALUES ($1, $2)
RETURNING created_at;
`, p.UUID, p.Username).
		Scan(&p.UUID, &p.Username, &p.CreatedAt)
}
