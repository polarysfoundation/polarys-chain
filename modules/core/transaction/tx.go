package transaction

import (
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
	Version   Version        `json:"version"`
	Payload   []byte         `json:"payload"`
}
