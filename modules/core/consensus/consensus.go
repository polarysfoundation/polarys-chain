package consensus

import (
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
)

type Engine interface {
	ConsensusProof(chainID uint64, crrBlockNumber uint64) ([]byte, error)
	ValidatorProof(block *block.Block) ([]byte, error)
	ValidatorExists(validator common.Address) bool
	VerifyBlock(block *block.Block) bool
	DifficultyValidator(block *block.Block) bool
	SealBlock(block *block.Block) (*block.Block, error)
	AdjustDifficulty(block *block.Block, prevBlock *block.Block) uint64
	SelectValidator() common.Address
}

type Chain interface {
	GetBlockByHash(hash common.Hash) (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlockByHeightAndHash(height uint64, hash common.Hash) (*block.Block, error)
	GetLatestBlock() (*block.Block, error)
}
