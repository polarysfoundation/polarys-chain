package accounts

import (
	"fmt"
	"sync"

	"github.com/polarysfoundation/polarys-chain/modules/accounts/keystore"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
)

type Accounts struct {
	accounts map[common.Address]*keystore.Wallet
	mutex    sync.RWMutex
}

func InitAccounts() *Accounts {
	accounts := keystore.GetLocalAccounts()

	w := &Accounts{
		accounts: make(map[common.Address]*keystore.Wallet),
	}

	if len(accounts) == 0 {
		return nil
	}

	for _, account := range accounts {
		wallet, err := keystore.InitWalletSecure(account)
		if err != nil {
			panic(err)
		}

		w.accounts[account] = wallet
	}

	return w
}

func (a *Accounts) NewAccount(passphrase []byte) (common.Address, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	w, err := keystore.NewWallet(passphrase)
	if err != nil {
		return common.Address{}, err
	}

	a.accounts[w.Address()] = w

	return w.Address(), nil
}

func (a *Accounts) Unlock(account common.Address, passphrase []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.accounts) == 0 {
		return nil
	}

	for _, acc := range a.accounts {
		if acc.Address() == account {
			return acc.Unlock(passphrase)
		}
	}

	return fmt.Errorf("account not found")
}

func (a *Accounts) Refresh() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.accounts) == 0 {
		return nil
	}

	for _, account := range a.accounts {
		err := account.Refresh()
		if err != nil {
			return err
		}
	}

	a.scan()

	return nil
}

func (a *Accounts) SignTX(account common.Address, tx *transaction.Transaction) (*transaction.Transaction, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if len(a.accounts) == 0 {
		return nil, nil
	}

	if wallet, ok := a.accounts[account]; ok {
		if wallet.IsLocked() {
			return nil, fmt.Errorf("account is locked")
		}

		return wallet.SignTX(tx)
	}

	return nil, fmt.Errorf("account not found")

}

func (a *Accounts) scan() {

	accounts := keystore.GetLocalAccounts()

	for _, account := range accounts {
		_, ok := a.accounts[account]
		if !ok {
			w, err := keystore.InitWalletSecure(account)
			if err != nil {
				panic(err)
			}

			a.accounts[account] = w
		}
	}
}
