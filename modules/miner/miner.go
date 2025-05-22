package miner

import (
	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
)

type wallet interface {
	Sign(a common.Address, data []byte) ([]byte, error)
	PubKey(a common.Address) (pec256.PubKey, error)
}

type Miner struct {
	address common.Address
	wallet  wallet
}

func NewMiner(address common.Address, wallet wallet) *Miner {
	return &Miner{
		address: address,
		wallet:  wallet,
	}
}

func (m *Miner) Address() common.Address {
	return m.address
}

func (m *Miner) PubKey() (pec256.PubKey, error) {
	return m.wallet.PubKey(m.address)
}

func (m *Miner) SignBlock(block *block.Block, chainID uint64) (*block.Block, error) {
	prefix := []byte{0xfb}

	b, err := common.Serialize([]interface{}{
		prefix,
		block.Nonce(),
		block.Timestamp(),
		block.Size(),
		block.Difficulty(),
		block.Height(),
		block.GasTarget(),
		chainID,
	})
	if err != nil {
		return nil, err
	}

	signature, err := m.wallet.Sign(m.address, b)
	if err != nil {
		return nil, err
	}

	return block.SignBlock(signature)
}
