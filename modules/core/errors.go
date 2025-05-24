package core

import "errors"

var (
	ErrBlockNotInitialized = errors.New("block not initialized")
	ErrBlockNotFound       = errors.New("block not found")
	ErrBlockExists         = errors.New("block already exists")
	ErrBlockHeight         = errors.New("invalid block height")
)
