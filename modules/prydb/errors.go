package prydb

import "errors"

var (
	ErrBlockNotFound        = errors.New("block not found")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrNotTransactionsFound = errors.New("no transactions found")
	ErrAccountNotFound      = errors.New("account not found")
	ErrTxPoolNotFound       = errors.New("tx pool not found")
)
