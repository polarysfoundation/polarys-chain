package txpool

import (
	"math/big"
	"sync"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/gaspool"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	"github.com/polarysfoundation/polarys-chain/modules/prydb"
)

type TxPool struct {
	poolAddress         common.Address
	executor            common.Address
	poolBalance         uint64
	timestamp           uint64
	consensusProof      []byte
	minimalGasTip       uint64
	gasProcessed        *big.Int
	nextEpoch           uint64
	pendingTransactions []transaction.Transaction
	sealedTransactions  []transaction.Transaction
	totalTransactions   uint64
	gaspool             *gaspool.GasPool
	latestBlock         *block.Block
	hash                common.Hash

	db    *prydb.Database
	mutex sync.RWMutex
}

func InitTxPool(db *prydb.Database, executor common.Address, minimalGasTip uint64, consensusProof []byte, gaspool *gaspool.GasPool, latestBlock *block.Block) (*TxPool, error) {

	h := crypto.Pm256(executor.Bytes())
	poolAddress := crypto.CreateAddress(executor, 0, common.BytesToHash(h))
	if db.TxPoolExist(poolAddress, latestBlock) {
		balance, timestamp, epoch, executor, hash, err := db.TxPoolState(poolAddress, latestBlock)
		if err != nil {
			return nil, err
		}

		pool := &TxPool{
			poolAddress:         poolAddress,
			executor:            executor,
			db:                  db,
			poolBalance:         balance,
			timestamp:           timestamp,
			nextEpoch:           epoch,
			consensusProof:      consensusProof,
			minimalGasTip:       minimalGasTip,
			gaspool:             gaspool,
			latestBlock:         latestBlock,
			hash:                hash,
			pendingTransactions: make([]transaction.Transaction, 0),
			sealedTransactions:  make([]transaction.Transaction, 0),
			gasProcessed:        big.NewInt(0),
			totalTransactions:   0,
		}

		return pool, nil

	}

	threeDaysEpoch := 3 * 24 * 60 * 60

	buff := make([]byte, 30)
	copy(buff[:15], executor.Bytes())
	copy(buff[15:], poolAddress.Bytes())

	h2 := crypto.Pm256(buff)

	pool := &TxPool{
		poolAddress:         poolAddress,
		executor:            executor,
		db:                  db,
		poolBalance:         0,
		timestamp:           uint64(time.Now().Unix()),
		pendingTransactions: make([]transaction.Transaction, 0),
		sealedTransactions:  make([]transaction.Transaction, 0),
		minimalGasTip:       minimalGasTip,
		gasProcessed:        big.NewInt(0),
		nextEpoch:           uint64(time.Now().Unix()) + uint64(threeDaysEpoch),
		consensusProof:      consensusProof,
		gaspool:             gaspool,
		hash:                common.BytesToHash(h2),
		totalTransactions:   0,
		latestBlock:         latestBlock,
	}

	return pool, nil
}

func (t *TxPool) PoolAddress() common.Address {
	return t.poolAddress
}

func (t *TxPool) AddTransaction(tx transaction.Transaction) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, existingTx := range t.pendingTransactions {
		if existingTx.Hash() == tx.Hash() {
			return ErrAlreadyExist
		}
	}

	t.pendingTransactions = append(t.pendingTransactions, tx)

	return nil
}

func (t *TxPool) GetTransactions() []transaction.Transaction {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.sealedTransactions
}

func (t *TxPool) ProcessTransaction() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for {
		seal := make([]byte, 96)
		h1 := crypto.Pm256(t.poolAddress.Bytes())
		h2 := crypto.Pm256(t.executor.Bytes())
		copy(seal[:32], h1)
		copy(seal[32:], h2)

		for _, tx := range t.pendingTransactions {
			copy(seal[64:], tx.Hash().Bytes())
			sealHash := crypto.Pm256(seal)
			tx.SealTx(common.BytesToHash(sealHash))

			t.sealedTransactions = append(t.sealedTransactions, tx)
		}

		t.pendingTransactions = make([]transaction.Transaction, 0)
	}
}

func (t *TxPool) Update(latestBlock *block.Block) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.latestBlock = latestBlock

	return nil
}

func (t *TxPool) AddBalance(amount uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.poolBalance += amount
}
