package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/golang-sql/sqlexp"
)

type Transaction interface {
	CreatePasswordResetCode(ctx context.Context, in authV1.PasswordResetCode) error
	CodeIsValid(ctx context.Context, code string) (bool, error)
	MarkAsUsed(ctx context.Context, code string) error
	MarkAllAsStale(ctx context.Context, accountID string) error

	GetPasswordForAccountID(ctx context.Context, id string) (*passwordV1.Password, error)
	CreatePassword(ctx context.Context, p passwordV1.Password) error
	UpdatePassword(ctx context.Context, p passwordV1.Password) error
}

type txImpl struct {
	tx *sql.Tx
}

func (tx *txImpl) CreatePasswordResetCode(ctx context.Context, in authV1.PasswordResetCode) error {
	var q sqlexp.Querier
	q = tx.tx

	c := entities.NewPasswordResetCodeFromPB(in)

	_, err := q.ExecContext(
		ctx,
		"INSERT INTO password_reset_code(code, expiration_timestamp, stale, used, account_id) VALUES($1, $2, $3, $4, $5)",
		c.Code, c.ExpiresAt, c.Stale, c.Used, c.AccountID)

	return err
}

func (tx *txImpl) CodeIsValid(ctx context.Context, code string) (bool, error) {
	var q sqlexp.Querier
	q = tx.tx

	var count int
	row := q.QueryRowContext(
		ctx,
		"SELECT COUNT(*) AS count "+
			"FROM password_reset_code "+
			"WHERE code = $1 "+
			"AND stale = FALSE "+
			"AND used = FALSE "+
			"AND expiration_timestamp > $2", code, time.Now())

	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 1, nil
}

func (tx *txImpl) MarkAsUsed(ctx context.Context, code string) error {
	var q sqlexp.Querier
	q = tx.tx

	_, err := q.ExecContext(
		ctx,
		"UPDATE password_reset_code SET used=TRUE, stale=TRUE WHERE code=$1",
		code)
	return err
}

func (tx *txImpl) MarkAllAsStale(ctx context.Context, accountID string) error {
	var q sqlexp.Querier
	q = tx.tx

	_, err := q.ExecContext(
		ctx,
		"UPDATE password_reset_code SET stale=TRUE WHERE account_id=$1",
		accountID)
	return err
}

func (tx *txImpl) GetPasswordForAccountID(ctx context.Context, id string) (*passwordV1.Password, error) {
	var q sqlexp.Querier
	q = tx.tx

	var p entities.Password
	row := q.QueryRowContext(
		ctx,
		"SELECT p.id, p.created_timestamp, p.iteration_count, p.salt, "+
			"p.password_hash, p.account_id "+
			"FROM Password p "+
			"WHERE p.id=$1", id)
	err := row.Scan(&p.Id, &p.Created, &p.IterationCount, &p.Salt, &p.PasswordHash, &p.AccountID)

	if err != nil {
		return nil, err
	}

	return p.ToProtobuf(), nil
}

func (tx *txImpl) CreatePassword(ctx context.Context, in passwordV1.Password) error {
	var q sqlexp.Querier
	q = tx.tx

	p := entities.NewPasswordFromProtobuf(in)

	_, err := q.ExecContext(
		ctx,
		"INSERT INTO Password(id, created_timestamp, iteration_count, salt, password_hash, account_id) VALUES($1, $2, $3, $4, $5, $6)",
		p.Id, p.Created, p.IterationCount, p.Salt, p.PasswordHash, p.AccountID)

	return err
}

func (tx *txImpl) UpdatePassword(ctx context.Context, in passwordV1.Password) error {
	var q sqlexp.Querier
	q = tx.tx

	p := entities.NewPasswordFromProtobuf(in)

	_, err := q.ExecContext(
		ctx,
		"UPDATE Password SET iteration_count=$1, salt=decode($2, 'hex'), password_hash=decode($3, 'hex') WHERE id=$4",
		p.IterationCount,
		hex.EncodeToString(p.Salt),
		hex.EncodeToString(p.PasswordHash),
		p.Id)
	return err
}
