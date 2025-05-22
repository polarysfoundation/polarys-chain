package keystore

import (
	"fmt"
	"sync"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/sirupsen/logrus"
)

type Wallet struct {
	address common.Address
	key     *Keypair
	locked  bool
	mutex   sync.RWMutex
	log     *logrus.Logger
}

func InitWalletSecure(a common.Address, log *logrus.Logger) (*Wallet, error) {
	w := &Wallet{
		address: a,
		key:     nil,
		locked:  true,
		log:     log,
	}
	return w, nil
}

func NewWallet(passphrase []byte, log *logrus.Logger) (*Wallet, error) {
	k, err := NewKeypair(passphrase)
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		address: k.address(),
		key:     k,
		locked:  false,
		log:     log,
	}
	return w, nil
}

func InitWalletWithKeypair(a common.Address, passphrase []byte) (*Wallet, error) {
	k, err := GetKeypairByAddress(a, passphrase)
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		address: a,
		key:     k,
		locked:  false,
	}
	return w, nil
}

func (w *Wallet) Address() common.Address {
	return w.address
}

func (w *Wallet) PubKey() pec256.PubKey {
	return w.key.pub
}

func (w *Wallet) SignTX(tx *transaction.Transaction) (*transaction.Transaction, error) {
	if w.IsLocked() {
		return nil, fmt.Errorf("wallet is locked")
	}

	return w.key.signTX(tx)
}

func (w *Wallet) Sign(data []byte) ([]byte, error) {
	if w.IsLocked() {
		return nil, fmt.Errorf("wallet is locked")
	}

	return w.key.sign(data)
}

func (w *Wallet) Refresh() error {
	if !w.key.expired() {
		return fmt.Errorf("keypair is not expired")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.key.lock()
	w.locked = true

	return nil
}

func (w *Wallet) IsLocked() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.locked
}

func (w *Wallet) Unlock(passphrase []byte) error {
	k, err := GetKeypairByAddress(w.address, passphrase)
	if err != nil {
		return err
	}

	if w.address.String() != k.address().String() {
		return fmt.Errorf("address not match, expected %s, got %s", w.address.String(), k.address().String())
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.key = k
	w.locked = false

	return nil
}
