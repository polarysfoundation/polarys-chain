package txpool

import "errors"

var (
	ErrNotFound     = errors.New("txpool not found")
	ErrAlreadyExist = errors.New("tx already exist")
)
