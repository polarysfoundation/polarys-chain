package consensus

import (
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

type Validator struct {
	Address      common.Address `json:"validator"`
	TotalReward  *big.Int       `json:"total_reward"`
	Staked       *big.Int       `json:"staked"`
	Power        *big.Int       `json:"power"`
	NextReward   *big.Int       `json:"next_reward"`
	CurrentEpoch uint64         `json:"current_epoch"`
	LastEpoch    uint64         `json:"last_epoch"`
}



