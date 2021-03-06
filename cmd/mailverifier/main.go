package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hhn-mc/mailverifier/internal/db"
	"github.com/hhn-mc/mailverifier/internal/mailer"
	"github.com/hhn-mc/mailverifier/internal/mailverifier"
	"github.com/hhn-mc/mailverifier/internal/player"
)

const configPathEnv = "MAILVERIFIER_CONFIG_PATH"

var configPath = "config.yaml"

func envString(name string, value string) string {
	envString := os.Getenv(name)
	if envString == "" {
		return value
	}

	return envString
}

func init() {
	configPath = envString(configPathEnv, configPath)

	if err := mailverifier.CreateConfigIfNotExist(configPath); err != nil {
		log.Fatalf("Failed to create default config at %s; %s", configPath, err)
	}
}

func main() {
	log.Printf("Reading config from %q", configPath)
	cfg, err := mailverifier.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
	}

	db := db.DB{
		Host:     cfg.Database.Host,
		Database: cfg.Database.Database,
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		Timeout:  10 * time.Second,
	}

	log.Printf("Opening connection to DB on %s with usernames %q",
		cfg.Database.Host, cfg.Database.Username)
	if err := db.Open(); err != nil {
		log.Fatalf("Failed connecting to the database; %s", err)
	}

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed mirgate the database schema; %s", err)
	}

	mailer := mailer.Service{
		Host:     cfg.Email.Host,
		SMTPHost: cfg.Email.SMTPHost,
		Email:    cfg.Email.Email,
		Alias:    cfg.Email.Alias,
		Identity: cfg.Email.Identity,
		Username: cfg.Email.Username,
		Password: cfg.Email.Password,
	}

	validityDuration, err := time.ParseDuration(cfg.EmailValidityDuration)
	if err != nil {
		log.Fatalf("Failed parse email validity duration; %s", err)
	}

	veCfg := player.VerificationEmailConfig{
		EmailRegex:             cfg.EmailRegex,
		VerificationCodeLength: cfg.VerificationCodeLength,
		EmailValidityDuration:  validityDuration,
		MaxEmailTries:          cfg.MaxEmailTries,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/players", func(r chi.Router) {
		r.Get("/{uuid}", player.GetPlayerHandler(&db))
		r.Post("/", player.PostPlayerHandler(&db))
		r.Route("/{uuid}/verifications", func(r chi.Router) {
			r.Use(player.ByUUIDMiddleware(&db))
			r.Get("/", player.GetVerificationsHandler(&db))
			r.Post("/", player.PostVerificationHandler(&db))
			r.Post("/verify", player.PostVerificationVerifyHandler(&db))
		})
		r.Route("/{uuid}/verification-emails", func(r chi.Router) {
			r.Use(player.ByUUIDMiddleware(&db))
			r.Post("/", player.PostVerificationEmailHandler(veCfg, mailer, &db))
		})
	})

	srv := http.Server{
		Addr:    cfg.API.Bind,
		Handler: r,
	}
	log.Printf("Starting API on %s", cfg.API.Bind)
	log.Println(srv.ListenAndServe())
}
