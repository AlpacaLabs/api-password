package entities

import (
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/google/uuid"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

type PasswordResetCode struct {
	ID        string
	Code      string
	Used      bool
	Stale     bool
	CreatedAt time.Time
	ExpiresAt time.Time
	AccountID string
}

func NewPasswordResetCode(accountID string, longevity time.Duration) (passwordV1.PasswordResetCode, error) {
	var empty passwordV1.PasswordResetCode
	var code string
	if u, err := uuid.NewRandom(); err != nil {
		return empty, err
	} else {
		code = u.String()
		log.Debugf("Generated password reset code: %s", code)
	}

	now := time.Now()

	return passwordV1.PasswordResetCode{
		Id:        xid.New().String(),
		AccountId: accountID,
		Code:      code,
		CreatedAt: clock.TimeToTimestamp(now),
		ExpiresAt: clock.TimeToTimestamp(now.Add(longevity)),
	}, nil
}

func NewPasswordResetCodeFromPB(c passwordV1.PasswordResetCode) PasswordResetCode {
	return PasswordResetCode{
		ID:        c.Id,
		Code:      c.Code,
		Used:      c.Used,
		Stale:     c.Stale,
		CreatedAt: clock.TimestampToTime(c.CreatedAt),
		ExpiresAt: clock.TimestampToTime(c.ExpiresAt),
		AccountID: c.AccountId,
	}
}

func (c PasswordResetCode) ToProtobuf() *passwordV1.PasswordResetCode {
	return &passwordV1.PasswordResetCode{
		Id:        c.ID,
		Code:      c.Code,
		Used:      c.Used,
		Stale:     c.Stale,
		CreatedAt: clock.TimeToTimestamp(c.CreatedAt),
		ExpiresAt: clock.TimeToTimestamp(c.ExpiresAt),
		AccountId: c.AccountID,
	}
}
