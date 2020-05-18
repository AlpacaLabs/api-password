package db

import (
	"context"
	"fmt"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"github.com/jackc/pgx/v4"
)

type PasswordResetTransaction interface {
	CreatePasswordResetCode(ctx context.Context, e entities.PasswordResetCode) error
	GetCode(ctx context.Context, codeID string) (*entities.PasswordResetCode, error)
	VerifyCode(ctx context.Context, code, accountID string) (*entities.PasswordResetCode, error)
	MarkAsUsed(ctx context.Context, codeID string) error
	MarkAllAsStale(ctx context.Context, accountID string) error
}

type passwordResetTransactionImpl struct {
	tx pgx.Tx
}

func (t *passwordResetTransactionImpl) CreatePasswordResetCode(ctx context.Context, c entities.PasswordResetCode) error {
	queryTemplate := `
INSERT INTO %s(id, code, created_at, expires_at, stale, used, account_id) 
 VALUES($1, $2, $3, $4, $5, $6, $7)
`

	query := fmt.Sprintf(queryTemplate, TableForPasswordResetCode)
	_, err := t.tx.Exec(ctx, query,
		c.ID, c.Code, c.CreatedAt, c.ExpiresAt, c.Stale, c.Used, c.AccountID)

	return err
}

func (t *passwordResetTransactionImpl) GetCode(ctx context.Context, codeID string) (*entities.PasswordResetCode, error) {
	queryTemplate := `
SELECT id, code, created_at, expires_at, stale, used, account_id
 FROM %s
 WHERE id = $1
`

	var c entities.PasswordResetCode

	query := fmt.Sprintf(queryTemplate, TableForPasswordResetCode)
	row := t.tx.QueryRow(ctx, query, codeID)

	err := row.Scan(&c.ID, &c.Code, &c.CreatedAt, &c.ExpiresAt, &c.Stale, &c.Used, &c.AccountID)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *passwordResetTransactionImpl) VerifyCode(ctx context.Context, code, accountID string) (*entities.PasswordResetCode, error) {
	queryTemplate := `
SELECT id, code, created_at, expires_at, stale, used, account_id
 FROM %s
 WHERE code = $1
 AND account_id = $2
 AND stale = FALSE
 AND used = FALSE
 AND expires_at > $3
`

	var c entities.PasswordResetCode

	query := fmt.Sprintf(queryTemplate, TableForPasswordResetCode)
	row := t.tx.QueryRow(ctx, query, code, accountID, time.Now())

	err := row.Scan(&c.ID, &c.Code, &c.CreatedAt, &c.ExpiresAt, &c.Stale, &c.Used, &c.AccountID)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *passwordResetTransactionImpl) MarkAsUsed(ctx context.Context, codeID string) error {
	queryTemplate := `
UPDATE %s 
 SET used=TRUE, stale=TRUE 
 WHERE id=$1
`

	query := fmt.Sprintf(queryTemplate, TableForPasswordResetCode)
	_, err := t.tx.Exec(ctx, query, codeID)
	return err
}

func (t *passwordResetTransactionImpl) MarkAllAsStale(ctx context.Context, accountID string) error {
	queryTemplate := `
UPDATE %s 
 SET stale=TRUE 
 WHERE account_id=$1
`

	query := fmt.Sprintf(queryTemplate, TableForPasswordResetCode)
	_, err := t.tx.Exec(ctx, query, accountID)
	return err
}
