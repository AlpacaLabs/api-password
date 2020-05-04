package entities

import (
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/guregu/null"
)

type PasswordResetCode struct {
	Code      string
	Used      bool
	Stale     bool
	ExpiresAt null.Time
	AccountID string
}

func NewPasswordResetCodeFromPB(c authV1.PasswordResetCode) PasswordResetCode {
	expiresAtTime := clock.TimestampToTime(c.ExpiresAt)
	var expiresAt null.Time
	if expiresAtTime.IsZero() {
		expiresAt = null.NewTime(time.Time{}, false)
	} else {
		expiresAt = null.TimeFrom(expiresAtTime)
	}
	return PasswordResetCode{
		Code:      c.Code,
		Used:      c.Used,
		Stale:     c.Stale,
		ExpiresAt: expiresAt,
		AccountID: c.AccountId,
	}
}

func (c PasswordResetCode) ToProtobuf() *authV1.PasswordResetCode {
	return &authV1.PasswordResetCode{
		Code:      c.Code,
		Used:      c.Used,
		Stale:     c.Stale,
		ExpiresAt: clock.TimeToTimestamp(c.ExpiresAt.ValueOrZero()),
		AccountId: c.AccountID,
	}
}
