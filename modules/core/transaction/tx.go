package transaction

import (
	"encoding/json"
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

type Version int

var (
	Legacy Version = 0
)

type TxData struct {
	From      common.Address `json:"from"`
	To        common.Address `json:"to"`
	Value     *big.Int       `json:"value"`
	Data      []byte         `json:"data"`
	Nonce     uint64         `json:"nonce"`
	Signature []byte         `json:"signature"`
	GasTip    uint64         `json:"gas_tip"`
	GasPrice  uint64         `json:"gas_price"`
	Gas       uint64         `json:"gas"`
	Version   Version        `json:"version"`
	Payload   []byte         `json:"payload"`
	Timestamp uint64         `json:"timestamp"`
}

func (t *TxData) Serialize() ([]byte, error) {
	temp := struct {
		From      common.Address `json:"from"`
		To        common.Address `json:"to"`
		Value     *big.Int       `json:"value"`
		Data      []byte         `json:"data"`
		Nonce     uint64         `json:"nonce"`
		GasTip    uint64         `json:"gas_tip"`
		GasPrice  uint64         `json:"gas_price"`
		Gas       uint64         `json:"gas"`
		Version   Version        `json:"version"`
		Payload   []byte         `json:"payload"`
		Timestamp uint64         `json:"timestamp"`
	}{
		From:      t.From,
		To:        t.To,
		Value:     t.Value,
		Data:      t.Data,
		Nonce:     t.Nonce,
		GasTip:    t.GasTip,
		GasPrice:  t.GasPrice,
		Gas:       t.Gas,
		Version:   t.Version,
		Payload:   t.Payload,
		Timestamp: t.Timestamp,
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil

}
