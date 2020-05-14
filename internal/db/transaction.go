package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
)

type Transaction interface {
	CreatePasswordResetCode(ctx context.Context, in passwordV1.PasswordResetCode) error
	GetCodeByCodeAndAccountID(ctx context.Context, code, accountID string) (*passwordV1.PasswordResetCode, error)
	MarkAsUsed(ctx context.Context, codeID string) error
	MarkAllAsStale(ctx context.Context, accountID string) error

	CreateTxobForCode(ctx context.Context, in passwordV1.DeliverCodeRequest) error

	GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error)
	CreatePassword(ctx context.Context, p passwordV1.Password) error
	UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error
}

type txImpl struct {
	tx pgx.Tx
}

func (tx *txImpl) CreatePasswordResetCode(ctx context.Context, in passwordV1.PasswordResetCode) error {
	c := entities.NewPasswordResetCodeFromPB(in)

	query := `
INSERT INTO password_reset_code(id, code, created_at, expires_at, stale, used, account_id) 
 VALUES($1, $2, $3, $4, $5, $6, $7)
`

	_, err := tx.tx.Exec(ctx, query,
		c.ID, c.Code, c.CreatedAt, c.ExpiresAt, c.Stale, c.Used, c.AccountID)

	return err
}

func (tx *txImpl) GetCodeByCodeAndAccountID(ctx context.Context, code, accountID string) (*passwordV1.PasswordResetCode, error) {
	query := `
SELECT id, code, created_at, expires_at, stale, used, account_id
 FROM password_reset_code
 WHERE code = $1
 AND account_id = $2
 AND stale = FALSE
 AND used = FALSE
 AND expires_at > $3
`

	var c entities.PasswordResetCode
	row := tx.tx.QueryRow(ctx, query, code, accountID, time.Now())

	err := row.Scan(&c.ID, &c.Code, &c.CreatedAt, &c.ExpiresAt, &c.Stale, &c.Used, &c.AccountID)
	if err != nil {
		return nil, err
	}

	return c.ToProtobuf(), nil
}

func (tx *txImpl) MarkAsUsed(ctx context.Context, codeID string) error {
	query := `
UPDATE password_reset_code 
 SET used=TRUE, stale=TRUE 
 WHERE id=$1
`

	_, err := tx.tx.Exec(ctx, query, codeID)
	return err
}

func (tx *txImpl) MarkAllAsStale(ctx context.Context, accountID string) error {
	query := `
UPDATE password_reset_code 
 SET stale=TRUE 
 WHERE account_id=$1
`

	_, err := tx.tx.Exec(ctx, query, accountID)
	return err
}

func (tx *txImpl) CreateTxobForCode(ctx context.Context, in passwordV1.DeliverCodeRequest) error {
	query := `
INSERT INTO password_reset_code_txob(code_id, sent, email_address_id, phone_number_id) 
 VALUES($1, FALSE, $2, $3)
`

	_, err := tx.tx.Exec(ctx, query, in.CodeId, in.GetEmailAddressId(), in.GetPhoneNumberId())

	return err
}

func (tx *txImpl) GetCurrentPasswordForAccountID(ctx context.Context, accountID string) (*passwordV1.Password, error) {
	query := `
SELECT p.id, p.created_at, p.iteration_count, p.salt, p.password_hash, p.account_id
 FROM password p
 JOIN account a
 ON a.id = p.account_id
 WHERE a.id=$1
`

	var p entities.Password
	row := tx.tx.QueryRow(ctx, query, accountID)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.IterationCount, &p.Salt, &p.Hash, &p.AccountID)

	if err != nil {
		return nil, err
	}

	return p.ToProtobuf(), nil
}

func (tx *txImpl) CreatePassword(ctx context.Context, in passwordV1.Password) error {
	query := `
INSERT INTO password(id, created_at, iteration_count, salt, password_hash, account_id) 
 VALUES($1, $2, $3, $4, $5, $6)
`

	p := entities.NewPasswordFromProtobuf(in)

	_, err := tx.tx.Exec(ctx, query, p.ID, p.CreatedAt, p.IterationCount, p.Salt, p.Hash, p.AccountID)

	return err
}

func (tx *txImpl) UpdateCurrentPassword(ctx context.Context, accountID, passwordID string) error {
	query := `
UPDATE account 
 SET current_password_id=$1, 
 WHERE id=$2
`

	_, err := tx.tx.Exec(ctx, query, passwordID, accountID)

	return err
}
