package db

import (
	"context"

	"github.com/AlpacaLabs/api-password/internal/db/entities"
	"github.com/jackc/pgx/v4"
)

type PasswordTransaction interface {
	GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (entities.Password, error)
	CreatePassword(ctx context.Context, p entities.Password) error
	UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error
}

type passwordTransactionImpl struct {
	tx pgx.Tx
}

func (t *passwordTransactionImpl) GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (entities.Password, error) {
	query := `
SELECT p.id, p.created_at, p.salt, p.password_hash, p.account_id, p.memory, p.iteration_count, p.parallelism, p.salt_length, p.key_length
  FROM password p
  JOIN account a
  ON a.id = p.account_id
  WHERE a.id=$1
`

	var p entities.Password

	row := t.tx.QueryRow(ctx, query, accountID)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.Salt, &p.Hash, &p.AccountID, &p.Memory, &p.Iterations, &p.Parallelism, &p.SaltLength, &p.KeyLength)

	if err != nil {
		return p, err
	}

	return p, nil
}

func (t *passwordTransactionImpl) CreatePassword(ctx context.Context, p entities.Password) error {
	query := `
INSERT INTO password(id, created_at, salt, password_hash, account_id, memory, iteration_count, parallelism, salt_length, key_length) 
  VALUES($1, $2, $3, $4, $5, $6)
`

	_, err := t.tx.Exec(ctx, query, p.ID, p.CreatedAt, p.Salt, p.Hash, p.AccountID, p.Memory, p.Iterations, p.Parallelism, p.SaltLength, p.KeyLength)

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
