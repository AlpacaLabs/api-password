package db

import (
	"github.com/jackc/pgx/v4"
)

const (
	TableForPasswordResetCode = "password_reset_code"
)

type Transaction interface {
	TransactionalOutbox
	PasswordTransaction
	PasswordResetTransaction
}

type txImpl struct {
	tx pgx.Tx
	outboxImpl
	passwordTransactionImpl
	passwordResetTransactionImpl
}

func newTransaction(tx pgx.Tx) Transaction {
	return &txImpl{
		tx: tx,
		outboxImpl: outboxImpl{
			tx: tx,
		},
		passwordTransactionImpl: passwordTransactionImpl{
			tx: tx,
		},
		passwordResetTransactionImpl: passwordResetTransactionImpl{
			tx: tx,
		},
	}
}
