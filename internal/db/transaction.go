package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/golang-sql/sqlexp"
)

type Transaction interface {
	CreatePasswordResetCode(ctx context.Context, in passwordV1.PasswordResetCode) error
	GetCodeByCodeAndAccountID(ctx context.Context, code, accountID string) (*passwordV1.PasswordResetCode, error)
	MarkAsUsed(ctx context.Context, code string) error
	MarkAllAsStale(ctx context.Context, accountID string) error

	CreateTxobForCode(ctx context.Context, codeID string) error

	GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error)
	CreatePassword(ctx context.Context, p passwordV1.Password) error
	UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error
}

type txImpl struct {
	tx *sql.Tx
}

func (tx *txImpl) CreatePasswordResetCode(ctx context.Context, in passwordV1.PasswordResetCode) error {
	var q sqlexp.Querier
	q = tx.tx

	c := entities.NewPasswordResetCodeFromPB(in)

	query := `
INSERT INTO password_reset_code(id, code, creation_timestamp, expiration_timestamp, stale, used, account_id) 
 VALUES($1, $2, $3, $4, $5)
`

	_, err := q.ExecContext(ctx, query, c.ID, c.Code, c.CreatedAt, c.ExpiresAt, c.Stale, c.Used, c.AccountID)

	return err
}

func (tx *txImpl) GetCodeByCodeAndAccountID(ctx context.Context, code, accountID string) (*passwordV1.PasswordResetCode, error) {
	var q sqlexp.Querier
	q = tx.tx

	query := `
SELECT id, code, creation_timestamp, expiration_timestamp, stale, used, account_id
 FROM password_reset_code
 WHERE code = $1
 AND account_id = $2
 AND stale = FALSE
 AND used = FALSE
 AND expiration_timestamp > $3
`

	var c entities.PasswordResetCode
	row := q.QueryRowContext(ctx, query, code, accountID, time.Now())

	err := row.Scan(&c.ID, &c.Code, &c.CreatedAt, &c.ExpiresAt, &c.Stale, &c.Used, &c.AccountID)
	if err != nil {
		return nil, err
	}

	return c.ToProtobuf(), nil
}

func (tx *txImpl) MarkAsUsed(ctx context.Context, code string) error {
	var q sqlexp.Querier
	q = tx.tx

	query := `
UPDATE password_reset_code 
 SET used=TRUE, stale=TRUE 
 WHERE code=$1
`

	_, err := q.ExecContext(ctx, query, code)
	return err
}

func (tx *txImpl) MarkAllAsStale(ctx context.Context, accountID string) error {
	var q sqlexp.Querier
	q = tx.tx

	query := `
UPDATE password_reset_code 
 SET stale=TRUE 
 WHERE account_id=$1
`

	_, err := q.ExecContext(ctx, query, accountID)
	return err
}

func (tx *txImpl) CreateTxobForCode(ctx context.Context, codeID string) error {
	var q sqlexp.Querier
	q = tx.tx

	query := `
INSERT INTO password_reset_code_txob(code_id, sent) 
 VALUES($1, FALSE)
`

	_, err := q.ExecContext(ctx, query, codeID)

	return err
}

func (tx *txImpl) GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error) {
	var q sqlexp.Querier
	q = tx.tx

	query := `
SELECT p.id, p.created_timestamp, p.iteration_count, p.salt, p.password_hash, p.account_id
 FROM password p
 JOIN account a
 ON a.id = p.account_id
 WHERE a.id=$1
`

	var p entities.Password
	row := q.QueryRowContext(ctx, query, accountID)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.IterationCount, &p.Salt, &p.Hash, &p.AccountID)

	if err != nil {
		return nil, err
	}

	return p.ToProtobuf(), nil
}

func (tx *txImpl) CreatePassword(ctx context.Context, in passwordV1.Password) error {
	var q sqlexp.Querier
	q = tx.tx

	query := `
INSERT INTO password(id, created_timestamp, iteration_count, salt, password_hash, account_id) 
 VALUES($1, $2, $3, $4, $5, $6)
`

	p := entities.NewPasswordFromProtobuf(in)

	_, err := q.ExecContext(ctx, query, p.ID, p.CreatedAt, p.IterationCount, p.Salt, p.Hash, p.AccountID)

	return err
}

func (tx *txImpl) UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error {
	var q sqlexp.Querier
	q = tx.tx

	query := `
UPDATE account 
 SET current_password_id=$1, 
 WHERE id=$2
`

	_, err := q.ExecContext(ctx, query, passwordID, accountID)

	return err
}
