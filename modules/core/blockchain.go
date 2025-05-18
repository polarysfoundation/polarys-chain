package core

import (
	"sync"

	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/core/txpool"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

type Blockchain struct {
	chainID        uint64
	localBlocks    []*block.Block
	txPool         txpool.TxPool
	consensusProof []byte
	epoch          uint64
	delay          uint64
	consensus      consensus.Engine
	latestBlock    *block.Block

	db   *polarysdb.Database
	lock sync.RWMutex
}

func InitBlockchain(db *polarysdb.Database, config *params.Config, chainParams *params.ChainParams) (*Blockchain, error) {
	bc := &Blockchain{
		chainID: params.Polarys.ChainID,
		epoch:   params.Polarys.PowEngine.Epoch,
		delay:   params.Polarys.PowEngine.Delay,
		db:      db,
	}

	if chainParams != nil {
		bc.chainID = chainParams.ChainID
		bc.epoch = chainParams.PowEngine.Epoch
		bc.delay = chainParams.PowEngine.Delay
	}

	return bc, nil
}
