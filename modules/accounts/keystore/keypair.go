package keystore

import (
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Keypair struct {
	pub       pec256.PubKey
	priv      pec256.PrivKey
	timestamp uint64
}

func (k *Keypair) close() {
	b := make([]byte, 0)

	k.priv = pec256.BytesToPrivKey(b)
	k.pub = pec256.BytesToPubKey(b)
}

func (k *Keypair) expired() bool {
	return uint64(time.Now().Unix()) > k.timestamp
}

func (k *Keypair) sharedKey() pec256.SharedKey {
	return crypto.GenerateSharedKey(k.priv)
}

func (k *Keypair) lock() {
	k.close()
	k.timestamp = 0
}

func (k *Keypair) address() common.Address {
	return crypto.PubKeyToAddress(k.pub)
}

func (k *Keypair) signTX(tx *transaction.Transaction) (*transaction.Transaction, error) {
	prefix := []byte{0xff}

	b, err := common.Serialize([]interface{}{
		prefix,
		k.sharedKey(),
		tx.Version(),
		tx.To(),
		tx.Nonce(),
		tx.Data(),
		tx.Value(),
		tx.Payload(),
	})
	if err != nil {
		return nil, err
	}

	h := crypto.Pm256(b)

	signature := make([]byte, 64)
	r, s, err := crypto.Sign(common.BytesToHash(h), k.priv)
	if err != nil {
		return nil, err
	}

	r.FillBytes(signature[32:])
	s.FillBytes(signature[:32])

	return tx.SignTransaction(signature), nil
}

func (k *Keypair) sign(data []byte) ([]byte, error) {
	h := crypto.Pm256(data)
	signature := make([]byte, 64)
	r, s, err := crypto.Sign(common.BytesToHash(h), k.priv)
	if err != nil {
		return nil, err
	}

	r.FillBytes(signature[32:])
	s.FillBytes(signature[:32])

	return signature, nil
}
