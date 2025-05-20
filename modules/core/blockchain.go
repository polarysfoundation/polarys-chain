package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/core/txpool"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

type Blockchain struct {
	chainID        uint64
	chainConfig    *params.Config
	genesis        block.Block
	localBlocks    []*block.Block
	txPool         *txpool.TxPool
	consensusProof []byte
	epoch          uint64
	delay          uint64
	consensus      consensus.Engine
	latestBlock    *block.Block

	db   *polarysdb.Database
	lock sync.RWMutex
}

func InitBlockchain(db *polarysdb.Database, config *params.Config, chainParams *params.ChainParams, engine consensus.Engine, genesis *GenesisBlock) (*Blockchain, error) {
	bc := &Blockchain{
		chainID:     chainParams.ChainID,
		epoch:       chainParams.PowEngine.Epoch,
		delay:       chainParams.PowEngine.Delay,
		chainConfig: config,
		db:          db,
		localBlocks: make([]*block.Block, 0),
	}

	if !hasGenesisBlock(db) {
		if genesis != nil {
			gb, err := InitGenesisBlock(db, false, genesis)
			if err != nil {
				return nil, err
			}

			blk, err := gb.ToBlock()
			if err != nil {
				return nil, err
			}

			bc.genesis = *blk
		} else {
			gb, err := InitGenesisBlock(db, true, nil)
			if err != nil {
				return nil, err
			}

			blk, err := gb.ToBlock()
			if err != nil {
				return nil, err
			}

			bc.genesis = *blk
		}

	}

	latestBlock, err := bc.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	bc.latestBlock = latestBlock

	consensusProof, err := engine.ConsensusProof(bc.chainID, latestBlock.Height())
	if err != nil {
		return nil, err
	}

	bc.consensus = engine
	bc.consensusProof = consensusProof

	txPool, err := txpool.InitTxPool(db, common.Address{}, uint64(config.MinimalGasTip), consensusProof)
	if err != nil {
		return nil, err
	}

	bc.txPool = txPool

	return bc, nil
}

func (bc *Blockchain) ChainID() uint64 {
	return bc.chainID
}

func (bc *Blockchain) ConsensusProof() []byte {
	return bc.consensusProof
}

func (bc *Blockchain) GetChainID() uint64 {
	return bc.chainID
}

func (bc *Blockchain) GetBlockByHash(hash common.Hash) (*block.Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return getBlockByHashAndHeight(bc.db, hash, 0)
}

func (bc *Blockchain) GetBlockByHeight(height uint64) (*block.Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return getBlockByHashAndHeight(bc.db, common.Hash{}, height)
}

func (bc *Blockchain) GetBlockByHeightAndHash(height uint64, hash common.Hash) (*block.Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return getBlockByHashAndHeight(bc.db, hash, height)
}

func (bc *Blockchain) GetLatestBlock() (*block.Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return getLatestBlock(bc.db)
}

func getLatestBlock(db *polarysdb.Database) (*block.Block, error) {
	data, ok := db.Read(metricCurrent, "0")
	if !ok {
		return nil, ErrBlockNotFound
	}

	var blk block.Block

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = blk.Deserialize(b, nil)
	if err != nil {
		return nil, err
	}

	txs, err := getTransactionsByBlockNumber(db, blk.Height())
	if err != nil {
		return nil, err
	}

	for _, tx := range txs {
		blk.AddTransaction(tx)
	}

	return &blk, nil
}

func getBlockByHashAndHeight(db *polarysdb.Database, hash common.Hash, height uint64) (*block.Block, error) {

	var data any
	var ok bool
	if !hash.IsValid() {
		data, ok = db.Read(metricByNumber, strconv.FormatUint(height, 10))
		if !ok {
			return nil, ErrBlockNotFound
		}
	} else {
		data, ok = db.Read(metricByHash, hash.String())
		if !ok {
			return nil, ErrBlockNotFound
		}
	}

	var blk block.Block

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	txs, err := getTransactionsByBlockNumber(db, height)
	if err != nil {
		return nil, err
	}

	err = blk.Deserialize(b, txs)
	if err != nil {
		return nil, err
	}

	return &blk, nil
}

func getTransactionsByBlockNumber(db *polarysdb.Database, height uint64) ([]transaction.Transaction, error) {
	data, err := db.ReadBatch(fmt.Sprintf(metricTransactionsByBlockNumber, strconv.FormatUint(height, 10)))
	if err != nil {
		return nil, err
	}

	txs := make([]transaction.Transaction, 0)

	for _, v := range data {
		var tx transaction.Transaction

		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		err = tx.Deserialize(b)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)

	}

	return txs, nil
}

func saveBlock(db *polarysdb.Database, blk *block.Block) error {
	err := db.Write(metricByNumber, strconv.FormatUint(blk.Height(), 10), blk)
	if err != nil {
		return err
	}

	err = db.Write(metricByHash, blk.Hash().String(), blk)
	if err != nil {
		return err
	}

	err = db.Write(metricCurrent, strconv.FormatUint(blk.Height(), 10), blk)
	if err != nil {
		return err
	}

	if len(blk.Transactions()) > 0 {
		for _, tx := range blk.Transactions() {
			err = db.Write(fmt.Sprintf(metricTransactionsByBlockNumber, strconv.FormatUint(blk.Height(), 10)), tx.Hash().String(), tx)
			if err != nil {
				return err
			}

			err = db.Write(fmt.Sprintf(metricTransactionsByBlockHash, blk.Hash().String()), tx.Hash().String(), tx)
			if err != nil {
				return err
			}
		}
	}

	return nil

}
