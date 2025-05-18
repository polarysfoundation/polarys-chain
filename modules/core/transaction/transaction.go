package transaction

import (
	"encoding/json"
	"math/big"
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	pm256 "github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Transaction struct {
	data     TxData
	hash     common.Hash
	sealHash common.Hash
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

func (t *Transaction) SignTx(priv pec256.PrivKey) error {
	b, err := t.data.Serialize()
	if err != nil {
		return err
	}

	h := crypto.Pm256(b)
	r, s, err := crypto.Sign(common.BytesToHash(h), priv)
	if err != nil {
		return err
	}

	signature := make([]byte, 64)
	copy(signature[:32], r.Bytes())
	copy(signature[32:], s.Bytes())

	t.data.Signature = signature

	return nil
}

func (t *Transaction) VerifyTx(pub pec256.PubKey) (bool, error) {
	b, err := t.data.Serialize()
	if err != nil {
		return false, err
	}

	h := crypto.Pm256(b)
	r := new(big.Int).SetBytes(t.data.Signature[:32])
	s := new(big.Int).SetBytes(t.data.Signature[32:])

	return crypto.Verify(common.BytesToHash(h), r, s, pub)
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
