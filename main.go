package main

import (
	"log"
	"time"
)

var configPath = "config.dev.yaml"

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
	}

	db := &database{
		dsn:     cfg.Database.dsn(),
		timeout: 10 * time.Second,
	}

	if err := db.open(); err != nil {
		log.Fatalf("Failed connecting to the database; %s", err)
	}

	if err := db.migrate(); err != nil {
		log.Fatalf("Failed mirgate the database schema; %s", err)
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
		db:                     db,
		emailRegex:             cfg.API.EmailRegex,
		verificationCodeLength: cfg.API.VerificationCodeLength,
	}

	log.Fatal(api.listenAndServe())
}
