package service

import (
	"context"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"github.com/AlpacaLabs/go-kontext"

	"github.com/AlpacaLabs/api-password/internal/db"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Service) DeliverCode(ctx context.Context, request passwordV1.DeliverCodeRequest) (*passwordV1.DeliverCodeResponse, error) {
	// TODO verify entity exists for this PK
	codeID := request.CodeId

	// TODO verify email address or phone number exists for this PK
	//emailAddressID := request.EmailAddressId

	if err := s.dbClient.RunInTransaction(ctx, func(ctx context.Context, tx db.Transaction) error {

		// Verify a Code entity exists for that primary key
		c, err := tx.GetCode(ctx, codeID)
		if err != nil {
			return err
		}

		if c.IsExpired() {
			return ErrCodeExpired
		}

		traceInfo := kontext.GetTraceInfo(ctx)

		return tx.CreateDeliverCodeRequest(ctx, entities.NewDeliverCodeRequest(traceInfo, request))
	}); err != nil {
		return nil, err
	}
	return &passwordV1.DeliverCodeResponse{}, nil
}
