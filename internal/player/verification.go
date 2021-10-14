package player

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Verification struct {
	ID         uint64              `json:"id"`
	PlayerUUID string              `json:"playerUuid,omitempty"`
	Emails     []VerificationEmail `json:"emails,omitempty"`
	IsVerified bool                `json:"isVerified"`
	CreatedAt  time.Time           `json:"createdAt"`
}

type VerificationEmail struct {
	Email      string     `json:"email"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (email VerificationEmail) Validate(emailRegex *regexp.Regexp) error {
	fieldRules := []*validation.FieldRules{
		validation.Field(&email.Email, validation.Required, validation.Match(emailRegex)),
	}

	return validation.ValidateStruct(&email, fieldRules...)
}

type VerificationEmailCode struct {
	Code string `json:"code"`
}

func (code VerificationEmailCode) Validate() error {
	fieldRules := []*validation.FieldRules{
		validation.Field(&code.Code, validation.Required),
	}

	return validation.ValidateStruct(&code, fieldRules...)
}

func generateVerificationCode(length int) (string, error) {
	bb := make([]byte, (length+1)/2)
	if _, err := rand.Read(bb); err != nil {
		return "", err
	}
	code := hex.EncodeToString(bb)[0:length]
	return strings.ToUpper(code), nil
}

type VerificationEmailConfig struct {
	EmailRegex             string
	VerificationCodeLength int
	EmailValidityDuration  time.Duration
	MaxEmailTries          int
}
