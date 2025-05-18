package txpool

import (
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

var (
	metric = "txpool/"
)

type TxPool struct {
	poolAddress         common.Address
	executor            common.Address
	poolBalance         *big.Int
	timestamp           uint64
	consensusProof      []byte
	minimalGasTip       uint64
	gasProcessed        *big.Int
	nextEpoch           uint64
	pendingTransactions []transaction.Transaction
	sealedTransactions  []transaction.Transaction
	totalTransactions   uint64

	db    *polarysdb.Database
	mutex sync.RWMutex
}

func (t *TxPool) Unmarshal(data []byte) error {
	temp := struct {
		PoolAddress       common.Address `json:"pool_address"`
		Executor          common.Address `json:"executor"`
		PoolBalance       *big.Int       `json:"pool_balance"`
		Timestamp         uint64         `json:"timestamp"`
		ConsensusProof    []byte         `json:"consensus_proof"`
		MinimalGasTip     uint64         `json:"minimal_gas_tip"`
		GasProcessed      *big.Int       `json:"gas_processed"`
		NextEpoch         uint64         `json:"next_epoch"`
		TotalTransactions uint64         `json:"total_transactions"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	t.poolAddress = temp.PoolAddress
	t.executor = temp.Executor
	t.poolBalance = temp.PoolBalance
	t.timestamp = temp.Timestamp
	t.consensusProof = temp.ConsensusProof
	t.minimalGasTip = temp.MinimalGasTip
	t.gasProcessed = temp.GasProcessed
	t.nextEpoch = temp.NextEpoch
	t.totalTransactions = temp.TotalTransactions

	return nil
}

func (t *TxPool) Marshal() ([]byte, error) {
	temp := struct {
		PoolAddress       common.Address `json:"pool_address"`
		Executor          common.Address `json:"executor"`
		PoolBalance       *big.Int       `json:"pool_balance"`
		Timestamp         uint64         `json:"timestamp"`
		ConsensusProof    []byte         `json:"consensus_proof"`
		MinimalGasTip     uint64         `json:"minimal_gas_tip"`
		GasProcessed      *big.Int       `json:"gas_processed"`
		NextEpoch         uint64         `json:"next_epoch"`
		TotalTransactions uint64         `json:"total_transactions"`
	}{
		PoolAddress:       t.poolAddress,
		Executor:          t.executor,
		PoolBalance:       t.poolBalance,
		Timestamp:         t.timestamp,
		ConsensusProof:    t.consensusProof,
		MinimalGasTip:     t.minimalGasTip,
		GasProcessed:      t.gasProcessed,
		NextEpoch:         t.nextEpoch,
		TotalTransactions: t.totalTransactions,
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func InitTxPool(db *polarysdb.Database, executor common.Address, minimalGasTip uint64, consensusProof []byte) (*TxPool, error) {
	if !db.Exist(metric) {
		err := db.Create(metric)
		if err != nil {
			return nil, err
		}
	}

	h := crypto.Pm256(executor.Bytes())
	poolAddress := crypto.CreateAddress(executor, 0, common.BytesToHash(h))
	if exist(db, poolAddress) {
		pool, err := getTxPool(db, poolAddress)
		if err != nil {
			return nil, err
		}

		return pool, nil
	}

	threeDaysEpoch := 3 * 24 * 60 * 60

	pool := &TxPool{
		poolAddress:         poolAddress,
		executor:            executor,
		db:                  db,
		poolBalance:         big.NewInt(0),
		timestamp:           uint64(time.Now().Unix()),
		pendingTransactions: make([]transaction.Transaction, 0),
		sealedTransactions:  make([]transaction.Transaction, 0),
		minimalGasTip:       minimalGasTip,
		gasProcessed:        big.NewInt(0),
		nextEpoch:           uint64(time.Now().Unix()) + uint64(threeDaysEpoch),
		consensusProof:      consensusProof,
	}

	err := save(db, pool)
	if err != nil {
		return nil, err
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

	seal := make([]byte, 96)
	h1 := crypto.Pm256(t.poolAddress.Bytes())
	h2 := crypto.Pm256(t.executor.Bytes())
	copy(seal[:32], h1)
	copy(seal[32:], h2)

	for _, tx := range t.pendingTransactions {
		h3 := crypto.Pm256(tx.Hash().Bytes())
		copy(seal[64:], h3)
		sealHash := crypto.Pm256(seal)
		tx.SealTx(common.BytesToHash(sealHash))

		t.sealedTransactions = append(t.sealedTransactions, tx)
	}
}

func (t *TxPool) Update() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	err := save(t.db, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *TxPool) AddBalance(amount *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.poolBalance.Add(t.poolBalance, amount)
}

func save(db *polarysdb.Database, pool *TxPool) error {
	err := db.Write(metric, pool.poolAddress.String(), pool)
	if err != nil {
		return err
	}

	return nil
}

func exist(db *polarysdb.Database, poolAddress common.Address) bool {
	_, ok := db.Read(metric, poolAddress.String())
	return ok
}

func getTxPool(db *polarysdb.Database, poolAddress common.Address) (*TxPool, error) {
	data, ok := db.Read(metric, poolAddress.String())
	if !ok {
		return nil, ErrNotFound
	}

	var txPool TxPool
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = txPool.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return &txPool, nil
}
