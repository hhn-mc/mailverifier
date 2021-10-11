package main

import (
	_ "embed"
)

type api struct {
	bind                   string
	mailer                 emailService
	emailRegex             string
	verificationCodeLength int
}

/*func (api api) listenAndServe() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/players/verifications", api.verificationsPostHandler())
	r.Post("/verifications/validate", api.verificationsValidateHandler())

	srv := http.Server{
		Addr:    api.bind,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func generateVerificationCode(length int) (string, error) {
	bb := make([]byte, length/2+1)
	if _, err := rand.Read(bb); err != nil {
		return "", err
	}
	code := hex.EncodeToString(bb)[0:length]
	return strings.ToUpper(code), nil
}

func (api api) verificationsPostHandler() http.HandlerFunc {
	emailPattern := regexp.MustCompile(api.emailRegex)
	return func(w http.ResponseWriter, r *http.Request) {
		var jsonReq struct {
			Username string `json:"username"`
			UUID     string `json:"uuid"`
			Email    string `json:"email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&jsonReq); err != nil {
			http.Error(w, "Failed to decode json", http.StatusBadRequest)
			return
		}

		username := jsonReq.Username
		if len(username) < 3 && len(username) > 16 {
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		uuid := jsonReq.UUID
		email := jsonReq.Email

		if email == "" || !emailPattern.MatchString(email) {
			http.Error(w, "Email in invalid format", http.StatusBadRequest)
			log.Printf("Email did not match: %q", email)
			return
		}

		api.db.createPlayerIfNotExist(uuid, username)

		verificationID, err := api.db.getUnverifiedValidationID(uuid)
		if err != nil {
			verificationID, err = api.db.createValidation(uuid)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				log.Printf("Failed to create verification in database; %s", err)
				return
			}
		}

		code, err := generateVerificationCode(api.verificationCodeLength)
		if err != nil {
			http.Error(w, "Failed to create verification code", http.StatusInternalServerError)
			log.Printf("Failed to create verification code; %s", err)
			return
		}

		if err := api.db.createEmailVerification(verificationID, code, email); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Failed to create email verification in database; %s", err)
			return
		}

		data := verificationEmailData{
			Code:     code,
			UUID:     uuid,
			Username: username,
			Time:     time.Now().Format(time.RFC3339),
			IP:       r.RemoteAddr,
		}
		if err := api.mailer.sendVerificationEmail(data, email); err != nil {
			http.Error(w, "Error while sending your email", http.StatusInternalServerError)
			log.Printf("Failed to send email to %q; %s", email, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (api api) verificationsValidateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jsonReq struct {
			UUID string `json:"uuid"`
			Code string `json:"code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&jsonReq); err != nil {
			http.Error(w, "Failed to decode json", http.StatusBadRequest)
			return
		}

		api.db.getActiveVerificationIDAndCodeForPlayerUUID(playerUUID)
	}
}*/
