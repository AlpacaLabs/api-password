package service

import (
	"context"

	"github.com/AlpacaLabs/api-password/internal/db"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Service) DeliverCode(ctx context.Context, request passwordV1.DeliverCodeRequest) (*passwordV1.DeliverCodeResponse, error) {
	if err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {
		return tx.CreateTxobForCode(ctx, request.CodeId)
	}); err != nil {
		return nil, err
	}
	return &passwordV1.DeliverCodeResponse{}, nil
}
