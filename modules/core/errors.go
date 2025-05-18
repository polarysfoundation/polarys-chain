package core

import "errors"

var (
	ErrBlockNotInitialized = errors.New("block not initialized")
	ErrBlockNotFound       = errors.New("block not found")
)
