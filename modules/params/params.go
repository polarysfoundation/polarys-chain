package params

import (
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

var (
	TestnetHash = common.StringToHash("")

	DefaultConfig = &Config{
		MaxProposalSize: 1024 * 1024,
		MaxTxSize:       1024 * 1024,
		MaxBlockSize:    1024 * 1024,
		MaxTxPerBlock:   1000,
	}

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

func (c *PowEngine) String() string {
	return "pow_engine"
}
