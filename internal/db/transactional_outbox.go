package db

import (
	"context"
	"fmt"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	"github.com/jackc/pgx/v4"
)

const (
	// TableForDeliverCodeRequest is the name of the transactional outbox (database table)
	// from which we read "jobs" or "events" that need to get sent to a message broker,
	TableForDeliverCodeRequest = "txob_deliver_code_request"
)

type TransactionalOutbox interface {
	CreateDeliverCodeRequest(ctx context.Context, request entities.DeliverCodeRequest) error
	ReadDeliverCodeRequest(ctx context.Context) (*entities.DeliverCodeRequest, error)
	MarkAsSentDeliverCodeRequest(ctx context.Context, eventID string) error
}

type outboxImpl struct {
	tx pgx.Tx
}

func (t *outboxImpl) CreateDeliverCodeRequest(ctx context.Context, in entities.DeliverCodeRequest) error {
	queryTemplate := `
INSERT INTO %s(event_id, trace_id, sampled, sent, code_id, email_address_id) 
 VALUES($1, $2, $3, $4, $5, $6)
`

	query := fmt.Sprintf(queryTemplate, TableForDeliverCodeRequest)
	_, err := t.tx.Exec(ctx, query, in.EventId, in.TraceId, in.Sampled, in.CodeId, false, in.GetEmailAddressId())

	return err
}

func (t *outboxImpl) ReadDeliverCodeRequest(ctx context.Context) (*entities.DeliverCodeRequest, error) {
	queryTemplate := `
SELECT event_id, trace_id, sampled, sent, code_id, email_address_id
  FROM %s
  WHERE sent = FALSE
  LIMIT 1
`

	query := fmt.Sprintf(queryTemplate, TableForDeliverCodeRequest)

	row := t.tx.QueryRow(ctx, query)

	var e entities.DeliverCodeRequest
	if err := row.Scan(&e.EventId, &e.TraceId, &e.Sampled, &e.Sent, &e.CodeId, &e.EmailAddressId); err != nil {
		return nil, err
	}

	return &e, nil
}

func (t *outboxImpl) MarkAsSentDeliverCodeRequest(ctx context.Context, eventID string) error {
	queryTemplate := `
UPDATE %s
  SET sent = TRUE
  WHERE event_id = $1
`
	query := fmt.Sprintf(queryTemplate, TableForDeliverCodeRequest)
	_, err := t.tx.Exec(ctx, query, eventID)
	return err
}
