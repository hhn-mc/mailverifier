package main

import "log"

var configPath = "config.dev.yaml"

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
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
		emailRegex: cfg.EmailRegex,
	}

	log.Fatal(api.listenAndServe())
}
