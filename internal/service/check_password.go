package service

import (
	"context"
	"errors"

	"github.com/AlpacaLabs/api-password/internal/db"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

var (
	ErrIncorrectCredentials = errors.New("either the account identifier or the password was incorrect")
)

func (s Service) CheckPassword(ctx context.Context, request passwordV1.CheckPasswordRequest) (*passwordV1.CheckPasswordResponse, error) {
	if err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		accountID := request.AccountId

		// Get current password
		currentPassword, err := tx.GetCurrentPasswordForAccountID(ctx, accountID)
		if err != nil {
			return err
		}

		// If the password is incorrect, return an error
		if !matchesHash(request.Password, currentPassword) {
			return ErrIncorrectCredentials
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return &passwordV1.CheckPasswordResponse{}, nil
}
