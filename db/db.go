package db

import (
	"fmt"
	"net"
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
	Host     string
	Database string
	Username string
	Password string
	Timeout  time.Duration

	*pgxpool.Pool
}

func (db DB) dsn() string {
	addr, port, _ := net.SplitHostPort(db.Host)
	return fmt.Sprintf(
		"host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode=disable",
		addr,
		port,
		db.Username,
		db.Password,
		db.Database,
	)
}

func (db *DB) Open() error {
	cfg, err := pgxpool.ParseConfig(db.dsn())
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

	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	db.Pool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}

	return db.Ping(ctx)
}

func (db *DB) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	_, err := db.Exec(ctx, schema)
	return err
}
