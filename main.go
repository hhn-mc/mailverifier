package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/hhn-mc/mailverifier/db"
	"github.com/hhn-mc/mailverifier/mailer"
	"github.com/hhn-mc/mailverifier/player"
	"github.com/hhn-mc/mailverifier/player/verification"
)

var configPath = "config.dev.yaml"

func main() {
	cfg, err := loadConfig(configPath)
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

	veCfg := verification.VerificationEmailConfig{
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
			r.Use(loadPlayer(db))
			r.Get("/", verification.GetVerificationsHandler(&db))
			r.Post("/", verification.PostVerificationHandler(&db))
		})
		r.Route("/{uuid}/verification-emails", func(r chi.Router) {
			r.Use(loadPlayer(db))
			r.Post("/", verification.PostVerificationEmailHandler(veCfg, mailer, &db))
		})
	})

	srv := http.Server{
		Addr:    cfg.API.Bind,
		Handler: r,
	}
	log.Println(srv.ListenAndServe())
}

func loadPlayer(db db.DB) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid := chi.URLParam(r, "uuid")
			if err := validation.Validate(uuid, is.UUIDv4); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			alreadyExists, err := db.PlayerWithUUIDExists(uuid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !alreadyExists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			player, err := db.PlayerByUUID(uuid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}

			ctx := context.WithValue(r.Context(), "player", player)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
