package entities

import (
	clocksql "github.com/AlpacaLabs/go-timestamp-sql"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/guregu/null"
)

// Password is a representation of a user's password.
type Password struct {
	Id             string    `json:"id"`
	Created        null.Time `json:"created_at"`
	IterationCount int
	Salt           []byte
	PasswordHash   []byte
	AccountID      string `json:"account_id"`
}

func (p Password) ToProtobuf() *passwordV1.Password {
	return &passwordV1.Password{
		Id:             p.Id,
		CreatedAt:      clocksql.TimestampFromNullTime(p.Created),
		IterationCount: int32(p.IterationCount),
		Salt:           p.Salt,
		Hash:           p.PasswordHash,
		AccountId:      p.AccountID,
	}
}

func NewPasswordFromProtobuf(p passwordV1.Password) Password {
	return Password{
		Id:             p.Id,
		Created:        clocksql.TimestampToNullTime(p.CreatedAt),
		IterationCount: int(p.IterationCount),
		Salt:           p.Salt,
		PasswordHash:   p.Hash,
		AccountID:      p.AccountId,
	}
}
