package prydb

import "encoding/json"

type account struct {
	nonce        uint64
	balance      uint64
	codeHash     []byte
	latestUpdate uint64
}

func InitAccount(nonce uint64, balance uint64, codeHash []byte, latestUpdate uint64) *account {
	return &account{nonce, balance, codeHash, latestUpdate}
}

func (a *account) GetNonce() uint64 {
	return a.nonce
}

func (a *account) GetBalance() uint64 {
	return a.balance
}

func (a *account) GetCodeHash() []byte {
	return a.codeHash
}

func (a *account) GetLatestUpdate() uint64 {
	return a.latestUpdate
}

func (a *account) SetNonce(nonce uint64) {
	a.nonce = nonce
}

func (a *account) SetBalance(balance uint64) {
	a.balance = balance
}

func (a *account) SetCodeHash(codeHash []byte) {
	a.codeHash = codeHash
}

func (a *account) SetLatestUpdate(latestUpdate uint64) {
	a.latestUpdate = latestUpdate
}

func (a *account) MarshalJSON() ([]byte, error) {
	temp := struct {
		Nonce        uint64 `json:"nonce"`
		Balance      uint64 `json:"balance"`
		CodeHash     []byte `json:"codeHash"`
		LatestUpdate uint64 `json:"latestUpdate"`
	}{
		Nonce:        a.nonce,
		Balance:      a.balance,
		CodeHash:     a.codeHash,
		LatestUpdate: a.latestUpdate,
	}
	return json.Marshal(temp)
}

func (a *account) UnmarshalJSON(data []byte) error {
	temp := struct {
		Nonce        uint64 `json:"nonce"`
		Balance      uint64 `json:"balance"`
		CodeHash     []byte `json:"codeHash"`
		LatestUpdate uint64 `json:"latestUpdate"`
	}{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	a.nonce = temp.Nonce
	a.balance = temp.Balance
	a.codeHash = temp.CodeHash
	a.latestUpdate = temp.LatestUpdate

	return nil

}
