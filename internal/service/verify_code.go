package service

import (
	"context"
	"errors"
	"strings"
	"time"

	clock "github.com/AlpacaLabs/go-timestamp"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/rs/xid"

	"github.com/AlpacaLabs/api-password/internal/db"
	"github.com/google/uuid"
)

var (
	ErrEmptyPassword = errors.New("password cannot be empty")
)

func (s *Service) VerifyCode(ctx context.Context, request passwordV1.VerifyCodeRequest) (*passwordV1.VerifyCodeResponse, error) {
	err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		if strings.TrimSpace(request.NewPassword) == "" {
			return ErrEmptyPassword
		}

		code := request.Code
		accountID := request.AccountId
		if _, err := uuid.Parse(request.Code); err != nil {
			return err
		}

		// Verify a code exists for the given code and account ID
		codeEntity, err := tx.VerifyCode(ctx, code, accountID)
		if err != nil {
			return err
		}

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

		if err := tx.MarkAsUsed(ctx, codeEntity.ID); err != nil {
			return err
		}

		if err := tx.MarkAllAsStale(ctx, accountID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &passwordV1.VerifyCodeResponse{}, nil
}
