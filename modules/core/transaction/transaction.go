package transaction

import (
	"encoding/json"
	"math/big"
	"time"

	pm256 "github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
)

type Transaction struct {
	data TxData
	hash common.Hash
	seal []byte
}

func NewTransaction(from common.Address, to common.Address, value *big.Int, data []byte, nonce uint64, gasTip uint64, gasPrice uint64, version Version, payload []byte) *Transaction {
	txData := TxData{
		From:      from,
		To:        to,
		Value:     value,
		Data:      data,
		Nonce:     nonce,
		GasTip:    gasTip,
		GasPrice:  gasPrice,
		Version:   version,
		Payload:   payload,
		Timestamp: uint64(time.Now().Unix()),
	}

	return &Transaction{
		data: txData,
	}
}

func (t *Transaction) Hash() common.Hash {
	if t.hash.IsValid() {
		return t.hash
	}

	data, err := json.Marshal(t.data)
	if err != nil {
		panic(err)
	}

	hash := pm256.Sum256(data)

	t.hash = common.BytesToHash(hash[:])
	return t.hash
}

func (t *Transaction) SealTx(seal []byte) {
	t.seal = seal
}

func (t *Transaction) Seal() []byte {
	return t.seal
}

func (t *Transaction) From() common.Address {
	return t.data.From
}

func (t *Transaction) To() common.Address {
	return t.data.To
}

func (t *Transaction) Value() *big.Int {
	return t.data.Value
}

func (t *Transaction) Data() []byte {
	return t.data.Data
}

func (t *Transaction) Nonce() uint64 {
	return t.data.Nonce
}

func (t *Transaction) Signature() []byte {
	return t.data.Signature
}

func (t *Transaction) GasTip() uint64 {
	return t.data.GasTip
}

func (t *Transaction) GasPrice() uint64 {
	return t.data.GasPrice
}

func (t *Transaction) Version() Version {
	return t.data.Version
}

func (t *Transaction) Payload() []byte {
	return t.data.Payload
}
