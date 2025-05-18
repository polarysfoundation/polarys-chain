package pow

import "errors"

var (
	ErrInvalidConsensusProof = errors.New("invalid consensus proof")
	ErrInvalidValidatorProof = errors.New("invalid validator proof")
	ErrInvalidValidator      = errors.New("invalid validator")
	ErrInvalidDifficulty     = errors.New("invalid difficulty")
	ErrInvalidBlockHeight    = errors.New("invalid block height")
	ErrInvalidBlockHash      = errors.New("invalid block hash")
	ErrInvalidEpoch          = errors.New("invalid epoch")
	ErrDuplicatedBlock       = errors.New("duplicated block")
	ErrInvalidSealHash       = errors.New("invalid seal hash")
)
