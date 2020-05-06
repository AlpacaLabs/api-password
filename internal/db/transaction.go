package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AlpacaLabs/api-password/internal/db/entities"

	authV1 "github.com/AlpacaLabs/protorepo-auth-go/alpacalabs/auth/v1"
	passwordV1 "github.com/AlpacaLabs/protorepo-password-go/alpacalabs/password/v1"
	"github.com/golang-sql/sqlexp"
	"go.opentelemetry.io/otel/api/global"
)

type Transaction interface {
	CreatePasswordResetCode(ctx context.Context, in authV1.PasswordResetCode) error

	GetPhoneNumbersForAccount(ctx context.Context, accountID string) ([]*authV1.PhoneNumber, error)
	GetPhoneNumber(ctx context.Context, phoneNumber string) (*authV1.PhoneNumber, error)

	GetConfirmedEmailAddressesForEmailAddress(ctx context.Context, emailAddress string) ([]*authV1.EmailAddress, error)
	GetConfirmedEmailAddressesForAccountID(ctx context.Context, accountID string) ([]*authV1.EmailAddress, error)

	CodeIsValid(ctx context.Context, code string) (bool, error)

	GetAccountIDForEmailAddress(ctx context.Context, emailAddress string) (string, error)
	GetAccountIDForUsername(ctx context.Context, username string) (string, error)

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

func (tx *txImpl) GetPhoneNumbersForAccount(ctx context.Context, accountID string) ([]*authV1.PhoneNumber, error) {
	// Start span
	tr := global.Tracer("password-reset")
	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	var q sqlexp.Querier
	q = tx.tx

	rows, err := q.QueryContext(
		ctx,
		"SELECT id, phone_number, account_id "+
			"FROM phone_number "+
			"WHERE confirmed=$1 AND account_id=$2 "+
			"AND deleted_timestamp IS NULL",
		true, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	phoneNumbers := []*authV1.PhoneNumber{}

	for rows.Next() {
		var p authV1.PhoneNumber
		if err := rows.Scan(&p.Id, &p.PhoneNumber, &p.AccountId); err != nil {
			return nil, err
		}

		phoneNumbers = append(phoneNumbers, &p)
	}

	return phoneNumbers, nil
}

func (tx *txImpl) GetPhoneNumber(ctx context.Context, phoneNumber string) (*authV1.PhoneNumber, error) {
	var q sqlexp.Querier
	q = tx.tx

	p := authV1.PhoneNumber{}

	err := q.QueryRowContext(
		ctx,
		"SELECT phone_number, account_id "+
			"FROM phone_number WHERE phone_number=$1 "+
			"AND deleted_timestamp IS NULL", phoneNumber).Scan(&p.PhoneNumber, &p.AccountId)

	if err != nil {
		// TODO don't log user-input
		return nil, fmt.Errorf("failed to read phone number: %s: %v", phoneNumber, err)
	}

	return &p, nil
}

func (tx *txImpl) GetConfirmedEmailAddressesForEmailAddress(ctx context.Context, emailAddress string) ([]*authV1.EmailAddress, error) {
	var q sqlexp.Querier
	q = tx.tx

	rows, err := q.QueryContext(
		ctx,
		"SELECT email_address, account_id "+
			"FROM email_address WHERE email_address=$1 "+
			"AND confirmed=$2 "+
			"AND deleted_timestamp IS NULL", emailAddress, true)

	if err != nil {
		return nil, fmt.Errorf("failed to find confirmed email addresses for account with email: %s: %v", emailAddress, err)
	}

	var out []*authV1.EmailAddress

	for {
		if !rows.Next() {
			break
		}
		entity := authV1.EmailAddress{
			EmailAddress: emailAddress,
		}
		if err := rows.Scan(&entity.EmailAddress, &entity.AccountId); err != nil {
			return nil, err
		}
		out = append(out, &entity)
	}

	return out, nil
}

func (tx *txImpl) GetConfirmedEmailAddressesForAccountID(ctx context.Context, accountID string) ([]*authV1.EmailAddress, error) {
	var q sqlexp.Querier
	q = tx.tx

	rows, err := q.QueryContext(
		ctx,
		"SELECT email_address, account_id "+
			"FROM email_address WHERE account_id=$1 "+
			"AND confirmed=$2 "+
			"AND deleted_timestamp IS NULL", accountID, true)

	if err != nil {
		return nil, fmt.Errorf("failed to find confirmed email addresses for account ID: %s: %v", accountID, err)
	}

	var out []*authV1.EmailAddress

	for {
		if !rows.Next() {
			break
		}
		entity := authV1.EmailAddress{
			AccountId: accountID,
		}
		if err := rows.Scan(&entity.EmailAddress, &entity.AccountId); err != nil {
			return nil, err
		}
		out = append(out, &entity)
	}

	return out, nil
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

func (tx *txImpl) GetAccountIDForEmailAddress(ctx context.Context, emailAddress string) (string, error) {
	var q sqlexp.Querier
	q = tx.tx

	var accountID string

	err := q.QueryRowContext(
		ctx,
		"SELECT id "+
			"FROM account WHERE email_address=$1 "+
			"AND deleted_timestamp IS NULL", emailAddress).Scan(&accountID)

	if err != nil {
		return "", err
	}

	// TODO if accountID is empty, return NotFound

	return accountID, nil
}

func (tx *txImpl) GetAccountIDForUsername(ctx context.Context, username string) (string, error) {
	var q sqlexp.Querier
	q = tx.tx

	var accountID string

	err := q.QueryRowContext(
		ctx,
		"SELECT id "+
			"FROM account WHERE username=$1 "+
			"AND deleted_timestamp IS NULL", username).Scan(&accountID)

	if err != nil {
		return "", err
	}

	// TODO if accountID is empty, return NotFound

	return accountID, nil
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
