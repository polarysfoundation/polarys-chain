package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/blockpool"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/core/gaspool"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/core/txpool"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	"github.com/polarysfoundation/polarys-chain/modules/prydb"
	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	chainID         uint64
	chainConfig     *params.Config
	genesis         block.Block
	localBlocks     []*block.Block
	txPool          *txpool.TxPool
	consensusProof  []byte
	epoch           uint64
	delay           uint64
	consensus       consensus.Engine
	latestBlock     *block.Block
	gasTarget       uint64
	difficulty      uint64
	totalDifficulty uint64
	blockPool       *blockpool.BlockPool
	ctx             context.Context
	wg              sync.WaitGroup
	cancel          context.CancelFunc
	gaspool         *gaspool.GasPool

	logs *logrus.Logger
	db   *prydb.Database
	lock sync.RWMutex
}

func InitBlockchain(db *prydb.Database, config *params.Config, chainParams *params.ChainParams, engine consensus.Engine, genesis *GenesisBlock, logs *logrus.Logger) (*Blockchain, error) {
	logs.WithFields(logrus.Fields{
		"chain_id": chainParams.ChainID,
	}).Info("Initializing blockchain")

	ctx, cancel := context.WithCancel(context.Background())

	bc := &Blockchain{
		chainID:         chainParams.ChainID,
		epoch:           chainParams.PowEngine.Epoch,
		delay:           chainParams.PowEngine.Delay,
		chainConfig:     config,
		db:              db,
		localBlocks:     make([]*block.Block, 0),
		difficulty:      chainParams.PowEngine.Difficulty,
		totalDifficulty: 0,
		logs:            logs,
		gasTarget:       1000000,
		ctx:             ctx,
		cancel:          cancel,
	}

	if !hasGenesisBlock(db) {
		bc.logs.Info("Genesis block not found, initializing genesis")

		if genesis != nil {
			gb, err := InitGenesisBlock(db, false, genesis, bc.consensusProof)
			if err != nil {
				bc.logs.WithError(err).Error("Failed to initialize provided genesis block")
				return nil, err
			}

			blk, err := gb.ToBlock()
			if err != nil {
				bc.logs.WithError(err).Error("Failed to convert genesis to block")
				return nil, err
			}

			blk.CalcHash()

			bc.logs.WithFields(logrus.Fields{
				"height": blk.Height(),
				"hash":   blk.Hash().String(),
			}).Info("Initializing genesis block")

			bc.genesis = *blk
		} else {
			gb, err := InitGenesisBlock(db, true, nil, bc.consensusProof)
			if err != nil {
				bc.logs.WithError(err).Error("Failed to create default genesis block")
				return nil, err
			}

			blk, err := gb.ToBlock()
			if err != nil {
				bc.logs.WithError(err).Error("Failed to convert default genesis to block")
				return nil, err
			}

			blk.CalcHash()

			bc.logs.WithFields(logrus.Fields{
				"height": blk.Height(),
				"hash":   blk.Hash().String(),
			}).Info("Initializing genesis block")

			bc.genesis = *blk
		}

		bc.logs.WithField("genesis_height", bc.genesis.Height()).Info("Genesis block initialized")
	} else {
		gb, err := getGenesisBlock(bc.db)
		if err != nil {
			bc.logs.WithError(err).Error("Failed to get genesis block")
			return nil, err
		}

		blk, err := gb.ToBlock()
		if err != nil {
			bc.logs.WithError(err).Error("Failed to get genesis block")
			return nil, err
		}

		blk.CalcHash()

		bc.genesis = *blk
	}

	latestBlock, err := bc.GetLatestBlock()
	if err != nil {
		bc.logs.WithError(err).Error("Failed to get latest block")
		return nil, err
	}

	if latestBlock.Height() == 0 {
		header := block.Header{
			Height:          1,
			Prev:            bc.genesis.Hash(),
			Nonce:           0,
			Timestamp:       uint64(time.Now().Unix()),
			GasTarget:       bc.gasTarget,
			Difficulty:      bc.difficulty,
			TotalDifficulty: bc.totalDifficulty,
			Validator:       common.Address{},
			ValidatorProof:  []byte{},
			ConsensusProof:  bc.consensusProof,
			Data:            []byte{},
			Signature:       []byte{},
			GasTip:          0,
			GasUsed:         0,
		}

		header.CalculateSize()
		blk := block.NewBlock(header, nil)
		blk.CalcHash()

		latestBlock = blk

		bc.latestBlock = blk
		bc.totalDifficulty += blk.Difficulty()

		bc.db.CommitBlock(latestBlock)
	}

	bc.latestBlock = latestBlock

	bc.logs.WithFields(logrus.Fields{
		"latest_height":    latestBlock.Height(),
		"latest_hash":      latestBlock.Hash().String(),
		"total_difficulty": bc.totalDifficulty,
	}).Info("Loaded latest block")

	consensusProof, err := engine.ConsensusProof(latestBlock.Height())
	if err != nil {
		bc.logs.WithError(err).Error("Failed to generate consensus proof")
		return nil, err
	}

	bc.gaspool = gaspool.InitGasPool()

	bc.consensus = engine
	bc.consensusProof = consensusProof

	txPool, err := txpool.InitTxPool(db, common.Address{}, uint64(config.MinimalGasTip), consensusProof, bc.gaspool, bc.latestBlock)
	if err != nil {
		bc.logs.WithError(err).Error("Failed to initialize transaction pool")
		return nil, err
	}

	blockPool, err := blockpool.NewBlockPool(engine, bc.db, latestBlock.Height(), config, bc.chainID, bc.epoch)
	if err != nil {
		bc.logs.WithError(err).Error("Failed to initialize block pool")
		return nil, err
	}

	bc.txPool = txPool

	bc.blockPool = blockPool

	bc.blockPool.SyncBlockPool(latestBlock.Height() + 1)

	bc.logs.WithField("tx_pool_initialized", true).Info("Blockchain initialized successfully")
	bc.logs.WithField("block_pool_initialized", true).Info("Blockchain initialized successfully")

	return bc, nil
}

func (bc *Blockchain) Difficulty() uint64 {
	return bc.difficulty
}

func (bc *Blockchain) ProtocolHash() common.Hash {
	return bc.consensus.ProtocolHash()
}

func (bc *Blockchain) GetTransactions() []transaction.Transaction {
	return bc.txPool.GetTransactions()
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

func (bc *Blockchain) AddBlock(blk *block.Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	bc.localBlocks = append(bc.localBlocks, blk)
	bc.latestBlock = blk

	return nil
}

func (bc *Blockchain) AddRemoteBlock(block *block.Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	if bc.hasBlock(block.Hash()) {
		return ErrBlockExists
	}

	if block.Height() <= bc.latestBlock.Height() {
		return ErrBlockHeight
	}

	err := bc.db.CommitBlock(block)
	if err != nil {
		return err
	}

	bc.latestBlock = block

	return nil
}

func (bc *Blockchain) HasBlock(hash common.Hash) bool {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.hasBlock(hash)
}

func (bc *Blockchain) hasBlock(hash common.Hash) bool {
	blk, err := bc.db.GetBlockByHash(hash)
	if err != nil {
		return false
	}

	return blk != nil
}

func (bc *Blockchain) Start() {
	bc.wg.Add(2)
	go bc.processLocalBlocksLoop()
	go bc.processBlocksLoop()
}

func (bc *Blockchain) processLocalBlocksLoop() {
	defer bc.wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bc.ctx.Done():
			bc.logs.Info("Stopping local-blocks loop")
			return
		case <-ticker.C:
			bc.lock.Lock()
			blocks := bc.localBlocks
			bc.lock.Unlock()
			if len(blocks) == 0 {
				continue
			}

			bc.logs.WithField("local_blocks", len(blocks)).Info("Processing local blocks")
			for i, blk := range blocks {
				if blk.Height() == bc.latestBlock.Height() {
					if err := bc.blockPool.AddProposedBlock(blk); err != nil {
						bc.logs.WithError(err).Error("Failed to add proposed block")
						continue
					}
					bc.lock.Lock()
					bc.localBlocks = append(bc.localBlocks[:i], bc.localBlocks[i+1:]...)
					bc.lock.Unlock()
					break
				}
			}
		}
	}
}

func (bc *Blockchain) processBlocksLoop() {
	defer bc.wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bc.ctx.Done():
			bc.logs.Info("Stopping block-pool loop")
			return
		case <-ticker.C:
			blk, err := bc.blockPool.ProcessProposedBlocks()
			if err != nil {
				continue
			}
			if blk == nil {
				continue
			}

			bc.logs.WithFields(logrus.Fields{
				"height": blk.Height(),
				"hash":   blk.Hash().String(),
			}).Info("Processing proposed block")

			// sólo guardamos si es la siguiente altura
			if blk.Height() == bc.latestBlock.Height() {
				latestBlock, err := bc.db.LatestBlock()
				if err != nil {
					bc.logs.WithError(err).Error("Failed to get latest block")
					continue
				}

				if err := bc.db.CommitBlock(blk); err != nil {
					bc.logs.WithError(err).Error("Failed to save new block")
					continue
				}

				bc.totalDifficulty += blk.Difficulty()
				bc.latestBlock = blk

				timeElapsed := time.Since(time.Unix(int64(latestBlock.Timestamp()), 0))

				bc.logs.WithFields(logrus.Fields{
					"height":           blk.Height(),
					"total_difficulty": bc.totalDifficulty,
					"delay":            fmt.Sprintf("%.2fs", float64(timeElapsed.Seconds())),
				}).Info("Committed new block")
			}

			if err := bc.blockPool.SyncBlockPool(blk.Height() + 1); err != nil {
				bc.logs.WithError(err).Error("Failed to sync block pool")
			}
		}
	}
}

func (bc *Blockchain) Stop() {
	bc.cancel()
	bc.wg.Wait()
	bc.logs.Info("Blockchain processing stopped")
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

	return bc.db.LatestBlock()
}

func getBlockByHashAndHeight(db *prydb.Database, hash common.Hash, height uint64) (*block.Block, error) {
	var err error
	var blk *block.Block

	if !hash.IsValid() {
		blk, err = db.GetBlockByHeight(height)
		if err != nil {
			return nil, err
		}

	} else {
		blk, err = db.GetBlockByHash(hash)
		if err != nil {
			return nil, err
		}
	}

	if ok, err := db.BlockHasTransactions(blk); err != nil {
		return nil, err
	} else if ok {
		txs, _ := db.GetTransactionsByBlockHeight(height)

		if len(txs) > 0 {
			for _, tx := range txs {
				blk.AddTransaction(*tx)
			}
		}

	}

	return blk, nil
}

func (bc *Blockchain) GasTarget() uint64 {
	return bc.gasTarget
}
