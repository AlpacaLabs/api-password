package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlpacaLabs/password-reset/internal/db"
	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
)

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrEmptyEmailAddress = errors.New("email address cannot be empty")
)

func (s *Service) ResetPassword(ctx context.Context, p authV1.ResetPasswordRequest) error {
	err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		if strings.TrimSpace(p.NewPassword) == "" {
			return ErrEmptyPassword
		}

		if strings.TrimSpace(p.EmailAddress) == "" {
			return ErrEmptyEmailAddress
		}

		if err := checkmail.ValidateFormat(p.EmailAddress); err != nil {
			return fmt.Errorf("invalid email: %v", err)
		}

		code := p.Code
		if _, err := uuid.Parse(p.Code); err != nil {
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

		accountID, err := tx.GetAccountIDForEmailAddress(ctx, p.EmailAddress)
		if err != nil {
			return err
		}

		// TODO hit auth service w/ account_id and new_password to reset password

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
