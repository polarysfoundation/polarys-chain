package main

import (
	"github.com/polarysfoundation/polarys-chain/modules/accounts"
	"github.com/polarysfoundation/polarys-chain/modules/miner"
)

func main() {
	accounts := accounts.InitAccounts()

	addr, err := accounts.NewAccount([]byte("test"))
	if err != nil {
		panic(err)
	}

	miner := miner.NewMiner(addr)
}
