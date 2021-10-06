package main

import "log"

var configPath = "config.dev.yaml"

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
	}

	db, err := openDatabase(cfg.Database.dsn())
	if err != nil {
		log.Fatalf("Failed connecting to the database; %s", err)
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
