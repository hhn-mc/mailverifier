package main

import (
	"log"
	"time"

	"github.com/hhn-mc/mailverifier/db"
)

var configPath = "config.dev.yaml"

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
	}

	db := &db.DB{
		DSN:          cfg.Database.dsn(),
		QueryTimeout: 10 * time.Second,
	}

	if err := db.Open(); err != nil {
		log.Fatalf("Failed connecting to the database; %s", err)
	}

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed mirgate the database schema; %s", err)
	}

	if err := db.PrepareStmts(); err != nil {
		log.Fatalf("Failed to prepare database statements; %s", err)
	}

	api := api{
		bind: cfg.API.Bind,
		mailer: emailService{
			host:     cfg.Email.Host,
			smtpHost: cfg.Email.SMTPHost,
			email:    cfg.Email.Email,
			identity: cfg.Email.Identity,
			username: cfg.Email.Username,
			password: cfg.Email.Password,
		},
		db:         db,
		emailRegex: cfg.API.EmailRegex,
		creds:      cfg.API.Creds,
	}

	log.Fatal(api.listenAndServe())
}
