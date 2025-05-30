package transaction

import (
	"encoding/json"
	"math/big"
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/gaspool"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Transaction struct {
	data     TxData
	hash     common.Hash
	sealHash common.Hash
}

func NewTransaction(from common.Address, to common.Address, value *big.Int, data []byte, nonce uint64, gasPrice uint64, version Version, payload []byte, gasTarget uint64) (*Transaction, error) {
	txData := &TxData{
		From:      from,
		To:        to,
		Value:     value,
		Data:      data,
		Nonce:     nonce,
		GasPrice:  gasPrice,
		Version:   version,
		Payload:   payload,
		Timestamp: uint64(time.Now().Unix()),
	}

	tx := &Transaction{
		data: *txData,
	}

	tx.CalcHash()

	tx, err := calcGas(gasTarget, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	temp := struct {
		TxData   TxData      `json:"tx_data"`
		Hash     common.Hash `json:"hash"`
		SealHash common.Hash `json:"seal_hash"`
	}{
		TxData:   t.data,
		Hash:     t.hash,
		SealHash: t.sealHash,
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxData   TxData      `json:"tx_data"`
		Hash     common.Hash `json:"hash"`
		SealHash common.Hash `json:"seal_hash"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	t.data = temp.TxData
	t.hash = temp.Hash
	t.sealHash = temp.SealHash

	return nil

}

func (t *Transaction) Hash() common.Hash {
	return t.hash
}

func (t *Transaction) Bytes() []byte {
	b, err := t.data.marshal()
	if err != nil {
		panic(err)
	}

	return b
}

func (t *Transaction) CalcHash() {
	data, err := t.data.marshal()
	if err != nil {
		panic(err)
	}

	hash := crypto.Pm256(data)

	t.hash = common.BytesToHash(hash)
}

func (t *Transaction) SignTransaction(signature []byte) *Transaction {
	auxTx := copyTransaction(t)
	auxTx.data.Signature = signature

	return auxTx
}

func (t *Transaction) VerifyTx(pub pec256.PubKey) (bool, error) {
	b, err := t.data.marshal()
	if err != nil {
		return false, err
	}

	h := crypto.Pm256(b)
	r := new(big.Int).SetBytes(t.data.Signature[:32])
	s := new(big.Int).SetBytes(t.data.Signature[32:])

	return crypto.Verify(common.BytesToHash(h), r, s, pub)
}

func (t *Transaction) Gas() uint64 {
	return t.data.Gas
}

func (t *Transaction) SealTx(seal common.Hash) {
	t.sealHash = seal
}

func (t *Transaction) Seal() common.Hash {
	return t.sealHash
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

func copyTransaction(tx *Transaction) *Transaction {
	return &Transaction{
		data:     tx.data,
		hash:     tx.hash,
		sealHash: tx.sealHash,
	}
}

func (t *Transaction) Timestamp() uint64 {
	return t.data.Timestamp
}

func (t *Transaction) CalcGas(gasTarget uint64) (*Transaction, error) {
	return calcGas(gasTarget, t)
}

func calcGas(gasTarget uint64, tx *Transaction) (*Transaction, error) {
	aux := copyTransaction(tx)

	payloadLen := uint64(len(aux.data.Payload))

	data, err := aux.data.marshal()
	if err != nil {
		return nil, err
	}

	dataLen := len(data)

	gasCost, gasTip, err := gaspool.CalcGas(gasTarget, tx.data.GasPrice, dataLen, int(payloadLen), tx.data.Value.BitLen())
	if err != nil {
		return nil, err
	}

	aux.data.Gas = gasCost
	aux.data.GasTip = gasTip

	return aux, err
}
