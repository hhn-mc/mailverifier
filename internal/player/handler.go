package player

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/hhn-mc/mailverifier/internal/mailer"
)

type DataRepo interface {
	PlayerWithUUIDExists(uuid string) (bool, error)
	PlayerByUUID(uuid string) (Player, error)
	CreatePlayer(p *Player) error

	Verifications(pUUID string) ([]Verification, error)
	CreateVerification(v *Verification) error
	LatestVerification(pUUID string) (Verification, bool, error)
	CreateEmailVerification(ve VerificationEmail) error
	VerifyVerification(vID uint64, code string) (bool, error)
}

func GetPlayerHandler(repo DataRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if err := validation.Validate(uuid, is.UUIDv4); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyExists, err := repo.PlayerWithUUIDExists(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !alreadyExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		player, err := repo.PlayerByUUID(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func PostPlayerHandler(repo DataRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var player Player
		if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := player.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyExists, err := repo.PlayerWithUUIDExists(player.UUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if alreadyExists {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if err := repo.CreatePlayer(&player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(player); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func GetVerificationsHandler(repo DataRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if err := validation.Validate(uuid, is.UUIDv4); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		verifications, err := repo.Verifications(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(verifications); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func PostVerificationHandler(repo DataRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if err := validation.Validate(uuid, is.UUIDv4); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		verification := Verification{
			PlayerUUID: uuid,
		}

		if err := repo.CreateVerification(&verification); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if err := json.NewEncoder(w).Encode(verification); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func PostVerificationEmailHandler(cfg VerificationEmailConfig, mail mailer.Service, repo DataRepo) http.HandlerFunc {
	emailRegex := regexp.MustCompile(cfg.EmailRegex)
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := r.Context().Value(CtxUUIDKey).(string)

		var email VerificationEmail
		if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := email.Validate(emailRegex); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		verification, exists, err := repo.LatestVerification(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed getting the latest verification; ", err)
			return
		}

		if !exists || verification.CreatedAt.Add(cfg.EmailValidityDuration).Before(time.Now()) {
			verification = Verification{PlayerUUID: uuid}
			if err := repo.CreateVerification(&verification); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("Failed creating a verification; ", err)
				return
			}
		}

		if len(verification.Emails) >= cfg.MaxEmailTries {
			http.Error(w, "Max email tries reached", http.StatusConflict)
			return
		}

		code, err := generateVerificationCode(cfg.VerificationCodeLength)
		if err != nil {
			http.Error(w, "Failed to create verification code", http.StatusInternalServerError)
			log.Printf("Failed to create verification code; %s", err)
			return
		}

		expiresAt := time.Now().Add(cfg.EmailValidityDuration)
		ve := VerificationEmail{
			VerificationID: verification.ID,
			Code:           code,
			Email:          email.Email,
			ExpiresAt:      &expiresAt,
		}
		if err := repo.CreateEmailVerification(ve); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed creating email verification; ", err)
			return
		}

		username := r.Context().Value(CtxUsernameKey).(string)
		emailData := mailer.VerificationEmailData{
			Code:     code,
			UUID:     uuid,
			Username: username,
			Time:     time.Now().Format(time.RFC3339),
		}
		if err := mail.SendVerificationEmail(emailData, email.Email); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed sending email", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func PostVerificationVerifyHandler(repo DataRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var code VerificationEmailCode
		if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := code.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		uuid := r.Context().Value(CtxUUIDKey).(string)
		validation, exists, err := repo.LatestVerification(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed getting the latest verification; ", err)
			return
		}

		if !exists {
			http.Error(w, "No pending verification", http.StatusBadRequest)
			return
		}

		success, err := repo.VerifyVerification(validation.ID, code.Code)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed verifying code; ", err)
			return
		}

		if !success {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
