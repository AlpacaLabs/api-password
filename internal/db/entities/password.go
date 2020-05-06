package entities

import (
	clocksql "github.com/AlpacaLabs/go-timestamp-sql"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/guregu/null"
)

// Password is a representation of a user's password.
type Password struct {
	ID             string
	CreatedAt      null.Time
	IterationCount int
	Salt           []byte
	Hash           []byte
	AccountID      string
}

func (p Password) ToProtobuf() *passwordV1.Password {
	return &passwordV1.Password{
		Id:             p.ID,
		CreatedAt:      clocksql.TimestampFromNullTime(p.CreatedAt),
		IterationCount: int32(p.IterationCount),
		Salt:           p.Salt,
		Hash:           p.Hash,
		AccountId:      p.AccountID,
	}
}

func NewPasswordFromProtobuf(p passwordV1.Password) Password {
	return Password{
		ID:             p.Id,
		CreatedAt:      clocksql.TimestampToNullTime(p.CreatedAt),
		IterationCount: int(p.IterationCount),
		Salt:           p.Salt,
		Hash:           p.Hash,
		AccountID:      p.AccountId,
	}
}
