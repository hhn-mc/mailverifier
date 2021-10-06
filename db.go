package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type database struct {
	*sql.DB
}

func openDatabase(dsn string) (database, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return database{}, err
	}

	if err := db.Ping(); err != nil {
		return database{}, err
	}

	return database{
		DB: db,
	}, err
}

func prepareStatements() {

}

func (db database) CreatePlayer(uuid string) error {
	return db.QueryRow(
		"INSERT INTO players (uuid) VALUES (UUID_TO_BIN(?));",
		uuid).Err()
}

func (db database) CreatePlayerEmailVerification(token, playerUUID string) error {
	return db.QueryRow(
		"INSERT INTO verifications (token, player_uuid) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?));",
		token, playerUUID).Err()
}

func (db database) InsertVerifiedEmail(token, email string) error {
	return db.QueryRow(
		"INSERT INTO verifications (email, verified_at) VALUES (?, CURRENT_TIMESTAMP) WHERE token = UUID_TO_BIN(?);",
		email, token).Err()
}
