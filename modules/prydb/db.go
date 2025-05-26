package prydb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

type Database struct {
	db             *polarysdb.Database
	cachedAccounts map[common.Address]*account
	cachedTxPools  map[common.Address]*txPool
}

func InitDB() (*Database, error) {
	db, err := polarysdb.Init(polarysdb.GenerateKeyFromBytes([]byte("test")), ".polarys")
	if err != nil {
		return nil, err
	}

	database := &Database{
		db:             db,
		cachedAccounts: make(map[common.Address]*account),
	}
	if err := database.initialize(); err != nil {
		return nil, err
	}

	return database, nil
}

func (db *Database) initialize() error {
	if !db.db.Exist(blocksByHash) {
		if err := db.db.Create(blocksByHash); err != nil {
			return err
		}
	}
	if !db.db.Exist(blocksByHeight) {
		if err := db.db.Create(blocksByHeight); err != nil {
			return err
		}
	}

	if !db.db.Exist(transactionsByHash) {
		if err := db.db.Create(transactionsByHash); err != nil {
			return err
		}
	}

	if !db.db.Exist(transactionsByAccount) {
		if err := db.db.Create(transactionsByAccount); err != nil {
			return err
		}
	}

	if !db.db.Exist(transactionsByBlockHash) {
		if err := db.db.Create(transactionsByBlockHash); err != nil {
			return err
		}
	}

	if !db.db.Exist(transactionsByBlockHeight) {
		if err := db.db.Create(transactionsByBlockHeight); err != nil {
			return err
		}
	}

	if !db.db.Exist(blocksLatest) {
		if err := db.db.Create(blocksLatest); err != nil {
			return err
		}
	}

	if !db.db.Exist(accounts) {
		if err := db.db.Create(accounts); err != nil {
			return err
		}
	}

	if !db.db.Exist(transactionsRejecteds) {
		if err := db.db.Create(transactionsRejecteds); err != nil {
			return err
		}
	}

	return nil

}

func (db *Database) CommitBlock(block *block.Block) error {
	return db.commitBlock(block)
}

func (db *Database) commitBlock(block *block.Block) error {
	if err := db.db.Write(blocksByHash, block.Hash().CXID(), block); err != nil {
		return err
	}

	if err := db.db.Write(blocksByHeight, strconv.FormatUint(block.Height(), 10), block); err != nil {
		return err
	}

	if err := db.db.Write(blocksLatest, "latest", block); err != nil {
		return err
	}

	return nil
}

func (db *Database) LatestBlock() (*block.Block, error) {
	data, ok := db.db.Read(blocksLatest, "latest")
	if !ok {
		return nil, nil
	}

	var block *block.Block
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &block); err != nil {
		return nil, err
	}

	return block, nil
}

func (db *Database) GetBlockByHash(hash common.Hash) (*block.Block, error) {
	data, ok := db.db.Read(blocksByHash, hash.CXID())
	if !ok {
		return nil, ErrBlockNotFound
	}

	var block *block.Block
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &block); err != nil {
		return nil, err
	}

	return block, nil

}

func (db *Database) GetBlockByHeight(height uint64) (*block.Block, error) {
	data, ok := db.db.Read(blocksByHash, strconv.FormatUint(height, 10))
	if !ok {
		return nil, ErrBlockNotFound
	}

	var block *block.Block
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &block); err != nil {
		return nil, err
	}

	return block, nil
}

func (db *Database) CommitTransaction(transaction *transaction.Transaction, block *block.Block) error {
	return db.commitTransaction(transaction, block)
}

func (db *Database) commitTransaction(transaction *transaction.Transaction, block *block.Block) error {
	if err := db.db.Write(transactionsByHash, transaction.Hash().CXID(), transaction); err != nil {
		return err
	}

	if err := db.db.Write(fmt.Sprintf(transactionsByAccount, transaction.From().CXID()), transaction.Hash().CXID(), transaction); err != nil {
		return err
	}

	if err := db.db.Write(fmt.Sprintf(transactionsByBlockHash, block.Hash().CXID()), transaction.Hash().CXID(), transaction); err != nil {
		return err
	}

	if err := db.db.Write(fmt.Sprintf(transactionsByBlockHeight, strconv.FormatUint(block.Height(), 10)), transaction.Hash().CXID(), transaction); err != nil {
		return err
	}

	return nil

}

func (db *Database) GetTransactionByHash(hash common.Hash) (*transaction.Transaction, error) {
	data, ok := db.db.Read(transactionsByHash, hash.CXID())
	if !ok {
		return nil, ErrTransactionNotFound
	}

	var transaction *transaction.Transaction

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (db *Database) GetTransactionsByAccount(account common.Address) ([]*transaction.Transaction, error) {
	data, ok := db.db.Read(transactionsByAccount, account.CXID())
	if !ok {
		return nil, ErrTransactionNotFound
	}

	var transactions []*transaction.Transaction

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil

}

func (db *Database) GetTransactionsByBlockHash(hash common.Hash) ([]*transaction.Transaction, error) {
	data, err := db.db.ReadBatch(fmt.Sprintf(transactionsByBlockHash, hash.CXID()))
	if err != nil {
		return nil, ErrTransactionNotFound
	}

	transactions := make([]*transaction.Transaction, 0)
	for _, v := range data {
		var transaction *transaction.Transaction
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &transaction)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)

	}

	return transactions, nil
}

func (db *Database) GetTransactionsByBlockHeight(height uint64) ([]*transaction.Transaction, error) {
	data, err := db.db.ReadBatch(fmt.Sprintf(transactionsByBlockHeight, strconv.FormatUint(height, 10)))
	if err != nil {
		return nil, ErrTransactionNotFound
	}

	transactions := make([]*transaction.Transaction, 0)
	for _, v := range data {
		var transaction *transaction.Transaction
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &transaction)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)

	}

	return transactions, nil
}

func (db *Database) GetTransactionRejectedByHash(hash common.Hash) (*transaction.Transaction, error) {
	data, ok := db.db.Read(transactionsRejecteds, hash.CXID())
	if !ok {
		return nil, ErrTransactionNotFound
	}

	var transaction *transaction.Transaction
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (db *Database) CommitTransactionRejected(transaction *transaction.Transaction) error {
	return db.commitTransactionRejected(transaction)
}

func (db *Database) commitTransactionRejected(transaction *transaction.Transaction) error {
	if err := db.db.Write(transactionsRejecteds, transaction.Hash().CXID(), transaction); err != nil {
		return err
	}

	return nil
}

func (db *Database) TransactionIsRejected(hash common.Hash) (bool, error) {
	_, ok := db.db.Read(transactionsRejecteds, hash.CXID())
	if !ok {
		return false, nil
	}

	return true, nil
}

func (db *Database) GetTransacionRejected() ([]*transaction.Transaction, error) {
	data, err := db.db.ReadBatch(transactionsRejecteds)
	if err != nil {
		return nil, ErrTransactionNotFound
	}

	if len(data) == 0 {
		return nil, ErrNotTransactionsFound
	}

	transactions := make([]*transaction.Transaction, 0)
	for _, v := range data {
		var transaction *transaction.Transaction
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &transaction)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)

	}

	return transactions, nil

}

func (db *Database) InitAccountState(address common.Address, code []byte, block *block.Block) error {
	account := InitAccount(0, 0, code, block.Height())

	return db.commitAccount(address, block, account)
}

func (db *Database) BalanceAt(address common.Address, block *block.Block) (uint64, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return 0, err
		}
	}

	acc, ok := db.cachedAccounts[address]
	if !ok {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return 0, err
		}

		db.cachedAccounts[address] = acc
	}

	if block.Height() != acc.latestUpdate {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return 0, err
		}

		db.cachedAccounts[address] = acc
	}

	return acc.balance, nil
}

func (db *Database) CodeAt(address common.Address, block *block.Block) ([]byte, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return nil, err
		}
	}

	acc, ok := db.cachedAccounts[address]
	if !ok {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return nil, err
		}

		db.cachedAccounts[address] = acc
	}

	if block.Height() != acc.latestUpdate {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return nil, err
		}

		db.cachedAccounts[address] = acc
	}

	return acc.codeHash, nil
}

func (db *Database) UpdateBalance(address common.Address, amount uint64, block *block.Block) error {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return err
		}
	}

	acc, ok := db.cachedAccounts[address]
	if !ok {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return err
		}

		db.cachedAccounts[address] = acc
	}

	if block.Height() != acc.latestUpdate {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return err
		}

		db.cachedAccounts[address] = acc
	}

	acc.balance = amount
	acc.latestUpdate = block.Height()

	return db.commitAccount(address, block, acc)

}

func (db *Database) getAccount(address common.Address, block *block.Block) (*account, error) {
	data, ok := db.db.Read(fmt.Sprintf(accounts, strconv.FormatUint(block.Height(), 10)), address.CXID())
	if !ok {
		return nil, ErrAccountNotFound
	}

	var acc *account
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &acc); err != nil {
		return nil, err
	}

	db.cachedAccounts[address] = acc

	return acc, nil
}

func (db *Database) commitAccount(address common.Address, block *block.Block, account *account) error {
	err := db.db.Write(fmt.Sprintf(accounts, strconv.FormatUint(block.Height(), 10)), address.CXID(), account)
	if err != nil {
		return err
	}

	return nil

}

func (db *Database) UpdateCode(address common.Address, code []byte, block *block.Block) error {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return err
		}
	}

	acc, ok := db.cachedAccounts[address]
	if !ok {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return err
		}

		db.cachedAccounts[address] = acc
	}

	if block.Height() != acc.latestUpdate {
		acc, err = db.getAccount(address, block)
		if err != nil {
			return err
		}

		db.cachedAccounts[address] = acc
	}

	acc.codeHash = code
	acc.latestUpdate = block.Height()

	return db.commitAccount(address, block, acc)
}

func (db *Database) TxPoolBalanceAt(address common.Address, block *block.Block) (uint64, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return 0, err
		}
	}

	txpool, ok := db.cachedTxPools[address]
	if !ok {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return 0, err
		}

		db.cachedTxPools[address] = txpool
	}

	if block.Height() != txpool.latestUpdate {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return 0, err
		}

		db.cachedTxPools[address] = txpool
	}

	return txpool.balance, nil
}

func (db *Database) TxPoolExcutor(address common.Address, block *block.Block) (common.Address, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return common.Address{}, err
		}
	}

	txpool, ok := db.cachedTxPools[address]
	if !ok {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return common.Address{}, err
		}

		db.cachedTxPools[address] = txpool
	}

	if block.Height() != txpool.latestUpdate {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return common.Address{}, err
		}

		db.cachedTxPools[address] = txpool
	}

	return txpool.executor, nil
}

func (db *Database) TxPoolHash(address common.Address, block *block.Block) (common.Hash, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return common.Hash{}, err
		}
	}

	txpool, ok := db.cachedTxPools[address]
	if !ok {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return common.Hash{}, err
		}

		db.cachedTxPools[address] = txpool
	}

	if block.Height() != txpool.latestUpdate {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return common.Hash{}, err
		}

		db.cachedTxPools[address] = txpool
	}

	return txpool.hash, nil
}

func (db *Database) getTxPool(address common.Address, block *block.Block) (*txPool, error) {
	data, ok := db.db.Read(fmt.Sprintf(txPools, strconv.FormatUint(block.Height(), 10)), address.CXID())
	if !ok {
		return nil, ErrTxPoolNotFound
	}

	var txpool *txPool
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &txpool); err != nil {
		return nil, err
	}

	db.cachedTxPools[address] = txpool

	return txpool, nil
}

func (db *Database) UpdateTxPoolBalance(address common.Address, amount uint64, block *block.Block) error {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return err
		}
	}

	acc, ok := db.cachedTxPools[address]
	if !ok {
		acc, err = db.getTxPool(address, block)
		if err != nil {
			return err
		}

		db.cachedTxPools[address] = acc
	}

	if block.Height() != acc.latestUpdate {
		acc, err = db.getTxPool(address, block)
		if err != nil {
			return err
		}

		db.cachedTxPools[address] = acc
	}

	acc.balance = amount
	acc.latestUpdate = block.Height()

	return db.commitTxPool(address, block, acc)

}

func (db *Database) TxPoolExist(address common.Address) bool {
	txpool, err := db.getTxPool(address, nil)
	if err != nil {
		return false
	}

	return txpool != nil
}

func (db *Database) commitTxPool(address common.Address, block *block.Block, txpool *txPool) error {
	err := db.db.Write(fmt.Sprintf(txPools, strconv.FormatUint(block.Height(), 10)), address.CXID(), txpool)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) TxPoolState(address common.Address, block *block.Block) (uint64, uint64, uint64, common.Address, common.Hash, error) {
	var err error
	if block == nil {
		block, err = db.LatestBlock()
		if err != nil {
			return 0, 0, 0, common.Address{}, common.Hash{}, err
		}
	}

	txpool, ok := db.cachedTxPools[address]
	if !ok {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return 0, 0, 0, common.Address{}, common.Hash{}, err

		}

		db.cachedTxPools[address] = txpool
	}

	if block.Height() != txpool.latestUpdate {
		txpool, err = db.getTxPool(address, block)
		if err != nil {
			return 0, 0, 0, common.Address{}, common.Hash{}, err
		}

		db.cachedTxPools[address] = txpool
	}

	return txpool.balance, txpool.timestamp, txpool.epoch, txpool.executor, txpool.hash, nil

}
