package service

import (
	"context"
	"fmt"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	hermesV1 "github.com/AlpacaLabs/protorepo-hermes-go/alpacalabs/hermes/v1"

	"github.com/AlpacaLabs/api-password/internal/db"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

func (s Service) DeliverCode(ctx context.Context, request *passwordV1.DeliverCodeRequest) (*passwordV1.DeliverCodeResponse, error) {
	funcName := "DeliverCode"
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

		payload := s.buildSendEmailRequest()
		transactionalOutboxTable := db.TableForSendEmailRequest

		// Create the event entity that will be persisted to the transactional outbox
		event, err := entities.NewEvent(ctx, request, payload)
		if err != nil {
			return fmt.Errorf("failed to create event in %s: %w", funcName, err)
		}

		// Persist the event to the transactional outbox
		return tx.CreateEvent(ctx, event, transactionalOutboxTable)

	}); err != nil {
		return nil, err
	}
	return &passwordV1.DeliverCodeResponse{}, nil
}

func (s Service) buildSendEmailRequest() *hermesV1.SendEmailRequest {
	// TODO build email
	return &hermesV1.SendEmailRequest{
		Email: nil,
	}
}
