package player

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Player struct {
	UUID       string    `json:"uuid"`
	Username   string    `json:"username"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (player Player) Validate() error {
	fieldRules := []*validation.FieldRules{
		validation.Field(&player.UUID, validation.Required, is.UUIDv4),
		validation.Field(&player.Username, validation.Required, validation.Length(1, 16), is.Alphanumeric),
	}

	return validation.ValidateStruct(&player, fieldRules...)
}

type Verification struct {
	ID         uint64    `json:"id"`
	PlayerUUID string    `json:"playerUuid,omitempty"`
	Email      string    `json:"email,omitempty"`
	VerifiedAt time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type VerificationsFilter struct {
	IsVerified *bool `schema:"isVerified"`
}
