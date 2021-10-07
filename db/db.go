package db

import (
	"database/sql"
	"sync"
	"time"

	_ "embed"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
)

//go:embed schema.sql
var schema string

type stmtID int

const (
	createPlayer stmtID = iota
	createVerification
	insertVerifiedEmail
)

type DB struct {
	DSN          string
	QueryTimeout time.Duration

	*sql.DB
	stmtsMu sync.Mutex
	stmts   map[stmtID]*sql.Stmt
}

func (db DB) Open() error {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.DB = sqlDB

	return db.Ping()
}

func (db DB) Migrate() error {
	row, err := db.Exec(schema)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) statement(id stmtID) *sql.Stmt {
	db.stmtsMu.Lock()
	defer db.stmtsMu.Unlock()
	return db.stmts[id]
}

func (db *DB) PrepareStmts() error {
	querys := map[stmtID]string{
		createPlayer:        "INSERT INTO players (uuid) VALUES (UUID_TO_BIN(?));",
		createVerification:  "INSERT INTO verifications (token, player_uuid) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?));",
		insertVerifiedEmail: "INSERT INTO verifications (email, verified_at) VALUES (?, CURRENT_TIMESTAMP) WHERE token = UUID_TO_BIN(?);",
	}

	db.stmtsMu.Lock()
	defer db.stmtsMu.Unlock()
	for id, query := range querys {
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		db.stmts[id] = stmt
	}
	return nil
}

func (db DB) transaction(t func(tx *sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.QueryTimeout)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := t(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db DB) CreatePlayerAndVerification(token, playerUUID string) error {
	return db.transaction(func(tx *sql.Tx) error {
		if _, err := tx.Stmt(db.statement(createPlayer)).Exec(playerUUID); err != nil {
			return err
		}

		if _, err := tx.Stmt(db.statement(createVerification)).Exec(token, playerUUID); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

func (db DB) insertVerifiedEmail(token, email string) error {
	return db.QueryRow(email, token).Err()
}
