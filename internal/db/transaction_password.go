package db

import (
	"context"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/jackc/pgx/v4"
)

type PasswordTransaction interface {
	GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error)
	CreatePassword(ctx context.Context, p passwordV1.Password) error
	UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error
}

type passwordTransactionImpl struct {
	tx pgx.Tx
}

func (t *passwordTransactionImpl) GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error) {
	query := `
SELECT p.id, p.created_at, p.iteration_count, p.salt, p.password_hash, p.account_id
 FROM password p
 JOIN account a
 ON a.id = p.account_id
 WHERE a.id=$1
`

	var p entities.Password

	row := t.tx.QueryRow(ctx, query, accountID)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.IterationCount, &p.Salt, &p.Hash, &p.AccountID)

	if err != nil {
		return nil, err
	}

	return p.ToProtobuf(), nil
}

func (t *passwordTransactionImpl) CreatePassword(ctx context.Context, in passwordV1.Password) error {
	query := `
INSERT INTO password(id, created_at, iteration_count, salt, password_hash, account_id) 
 VALUES($1, $2, $3, $4, $5, $6)
`

	p := entities.NewPasswordFromProtobuf(in)

	_, err := t.tx.Exec(ctx, query, p.ID, p.CreatedAt, p.IterationCount, p.Salt, p.Hash, p.AccountID)

	return err
}

func (t *passwordTransactionImpl) UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error {
	query := `
UPDATE account 
 SET current_password_id=$1, 
 WHERE id=$2
`

	_, err := t.tx.Exec(ctx, query, passwordID, accountID)

	return err
}
