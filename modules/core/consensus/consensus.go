package consensus

import (
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
)

var (
	SystemAddress = common.CXIDToAddress("1cxffffFFFfFFffffffffffffffFfFFFfffFFFfFFfE")
)

type Engine interface {
	ConsensusProof(crrBlockNumber uint64) ([]byte, error)
	ValidatorProof() ([]byte, error)
	ValidatorExists(validator common.Address) bool
	VerifyBlock(chain Chain, block *block.Block) (bool, error)
	DifficultyValidator(block *block.Block, prevBlock *block.Block) bool
	SealBlock(block *block.Block) (*block.Block, error)
	AdjustDifficulty(block *block.Block, prevBlock *block.Block) uint64
	SelectValidator() common.Address
	Validator() common.Address
	VerifyChain(chain Chain) (bool, error)
}

type Chain interface {
	GetBlockByHash(hash common.Hash) (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlockByHeightAndHash(height uint64, hash common.Hash) (*block.Block, error)
	GetLatestBlock() (*block.Block, error)
}
