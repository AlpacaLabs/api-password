package service

import (
	"context"
	"errors"

	"github.com/AlpacaLabs/password-reset/internal/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (s *Service) VerifyCode(ctx context.Context, codeString string) (bool, error) {

	log.Debugf("Verifying password reset code: %s", codeString)

	// Validate the password reset code is a UUID
	if _, err := uuid.Parse(codeString); err != nil {
		return false, err
	}

	var validCode bool

	err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		b, err := tx.CodeIsValid(ctx, codeString)
		if err != nil {
			return err
		}
		validCode = b
		return nil
	})
	if err != nil {
		return false, err
	}

	if !validCode {
		return false, errors.New("reset code not valid")
	}

	return true, nil
}
