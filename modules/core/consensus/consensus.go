package consensus

import (
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
)

type Engine interface {
	InitConsensus(epoch, difficulty, delay uint64) error
	ConsensusProof(chainID uint64, crrBlockNumber uint64) ([]byte, error)
	VerifyBlock(block *block.Block) bool
	ValidatorExists(validator common.Address) bool
	DifficultyValidator(block *block.Block) bool
}

type Chain interface {
	GetBlockByHash(hash common.Hash) (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlockByHeightAndHash(height uint64, hash common.Hash) (*block.Block, error)
	GetLatestBlock() (*block.Block, error)
}
