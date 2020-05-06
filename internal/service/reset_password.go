package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/rs/xid"

	"github.com/AlpacaLabs/api-password/internal/db"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
)

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrEmptyEmailAddress = errors.New("email address cannot be empty")
)

func (s *Service) ResetPassword(ctx context.Context, request authV1.ResetPasswordRequest) error {
	err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		if strings.TrimSpace(request.NewPassword) == "" {
			return ErrEmptyPassword
		}

		if strings.TrimSpace(request.EmailAddress) == "" {
			return ErrEmptyEmailAddress
		}

		if err := checkmail.ValidateFormat(request.EmailAddress); err != nil {
			return fmt.Errorf("invalid email: %v", err)
		}

		code := request.Code
		if _, err := uuid.Parse(request.Code); err != nil {
			return err
		}

		if valid, err := tx.CodeIsValid(ctx, code); err != nil {
			return err
		} else if !valid {
			// The password reset code is not valid,
			// but we don't want to leak to the client.
			// Clients should display something like:
			// "If the reset code was valid, your password was reset."
			return nil
		}

		account, err := s.getAccountForEmailAddress(ctx, request.EmailAddress)
		if err != nil {
			return err
		}

		accountID := account.Id

		salt, err := generateSalt(32)
		if err != nil {
			return err
		}
		iterationCount := 10000
		hash := generateHash(request.NewPassword, iterationCount, salt)

		newPasswordID := xid.New().String()
		if err := tx.CreatePassword(ctx, passwordV1.Password{
			Id:             newPasswordID,
			CreatedAt:      clock.TimeToTimestamp(time.Now()),
			IterationCount: int32(iterationCount),
			Salt:           salt,
			Hash:           hash,
			AccountId:      accountID,
		}); err != nil {
			return err
		}

		if err := tx.UpdateCurrentPassword(ctx, accountID, newPasswordID); err != nil {
			return err
		}

		if err := tx.MarkAsUsed(ctx, code); err != nil {
			return err
		}

		if err := tx.MarkAllAsStale(ctx, accountID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
