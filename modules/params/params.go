package params

import (
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

var (
	TestnetHash = common.StringToHash("")

	Polarys = &ChainParams{
		PolarysBlock: big.NewInt(0),
		ChainID:      0,
		PowEngine: PowEngine{
			Epoch:      1000,
			Difficulty: 100,
			Delay:      10,
		},
	}
)

type ChainParams struct {
	PolarysBlock *big.Int
	ChainID      uint64
	PowEngine    PowEngine
}

type PowEngine struct {
	Epoch      uint64
	Difficulty uint64
	Delay      uint64
}
