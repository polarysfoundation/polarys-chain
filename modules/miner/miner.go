package miner

import (
	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
)

type Miner struct {
	address common.Address
	pubKey  pec256.PubKey
	privKey pec256.PrivKey
}

func NewMiner(address common.Address, pubKey pec256.PubKey, privKey pec256.PrivKey) *Miner {
	return &Miner{
		address: address,
		pubKey:  pubKey,
		privKey: privKey,
	}
}

func (m *Miner) Address() common.Address {
	return m.address
}

func (m *Miner) PubKey() pec256.PubKey {
	return m.pubKey
}

func (m *Miner) PrivKey() pec256.PrivKey {
	return m.privKey
}
