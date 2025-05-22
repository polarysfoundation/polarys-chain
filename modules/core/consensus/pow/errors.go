package pow

import "errors"

// Define error variables
var (
	ErrInvalidBlockHash      = errors.New("invalid block hash")
	ErrInvalidConsensusProof = errors.New("invalid consensus proof")
	ErrInvalidValidatorProof = errors.New("invalid validator proof")
	ErrInvalidValidator      = errors.New("invalid validator")
	ErrInvalidBlockHeight    = errors.New("invalid block height")
	ErrInvalidEpoch          = errors.New("invalid epoch")
	ErrDuplicatedBlock       = errors.New("duplicated block")
	ErrInvalidSealHash       = errors.New("invalid seal hash")
	ErrInvalidDifficulty     = errors.New("invalid difficulty")
	ErrNilBlock              = errors.New("block is nil")
	ErrNilPreviousBlock      = errors.New("previous block is nil")
	ErrInvalidBlockSize      = errors.New("invalid block size")
	ErrInvalidBlockTimestamp = errors.New("invalid block timestamp")
	ErrInvalidBlockNonce     = errors.New("invalid block nonce")
	ErrInvalidValidatorCount = errors.New("invalid validator count")
	ErrInvalidProtocolHash   = errors.New("invalid protocol hash")
	ErrInvalidChainID        = errors.New("invalid chain ID")
)
