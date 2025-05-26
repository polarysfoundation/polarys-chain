package prydb

import (
	"encoding/json"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

type txPool struct {
	balance      uint64
	hash         common.Hash
	timestamp    uint64
	epoch        uint64
	executor     common.Address
	latestUpdate uint64
}

func InitTxPool(balance uint64, hash common.Hash, timestamp uint64, epoch uint64, executor common.Address, lastUpdate uint64) *txPool {
	return &txPool{balance, hash, timestamp, epoch, executor, lastUpdate}
}

func (tx *txPool) MarshalJSON() ([]byte, error) {
	tmp := struct {
		Balance      uint64         `json:"balance"`
		Hash         common.Hash    `json:"hash"`
		Timestamp    uint64         `json:"timestamp"`
		Epoch        uint64         `json:"epoch"`
		Executor     common.Address `json:"executor"`
		LatestUpdate uint64         `json:"last_update"`
	}{
		Balance:      tx.balance,
		Hash:         tx.hash,
		Timestamp:    tx.timestamp,
		Epoch:        tx.epoch,
		Executor:     tx.executor,
		LatestUpdate: tx.latestUpdate,
	}

	return json.Marshal(tmp)
}

func (tx *txPool) UnmarshalJSON(data []byte) error {
	tmp := struct {
		Balance      uint64         `json:"balance"`
		Hash         common.Hash    `json:"hash"`
		Timestamp    uint64         `json:"timestamp"`
		Epoch        uint64         `json:"epoch"`
		Executor     common.Address `json:"executor"`
		LatestUpdate uint64         `json:"last_update"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	tx.balance = tmp.Balance
	tx.hash = tmp.Hash
	tx.timestamp = tmp.Timestamp
	tx.epoch = tmp.Epoch
	tx.executor = tmp.Executor
	tx.latestUpdate = tmp.LatestUpdate

	return nil

}
