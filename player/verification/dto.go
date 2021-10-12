package verification

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Verification struct {
	ID         uint64    `json:"id"`
	PlayerUUID string    `json:"playerUuid,omitempty"`
	Emails     []Email   `json:"emails,omitempty"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Email struct {
	Email      string     `json:"email"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (email Email) Validate(emailRegex *regexp.Regexp) error {
	fieldRules := []*validation.FieldRules{
		validation.Field(&email.Email, validation.Required, validation.Match(emailRegex)),
	}

	return validation.ValidateStruct(&email, fieldRules...)
}
