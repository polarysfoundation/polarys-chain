package main

import (
	"github.com/polarysfoundation/polarys-chain/modules/accounts"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus/pow"
	"github.com/polarysfoundation/polarys-chain/modules/miner"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	polarysdb "github.com/polarysfoundation/polarys_db"
	"github.com/sirupsen/logrus"
)

func main() {
	/* 	addr := common.CXIDToAddress("1cx7fa5a303d068119a5ab0d500daf0ba") */

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	accounts := accounts.InitAccounts(logger)

	addr, err := accounts.NewAccount([]byte("test"))
	if err != nil {
		panic(err)
	}

	m := miner.NewMiner(addr, accounts)

	db, err := polarysdb.Init(polarysdb.GenerateKeyFromBytes([]byte("test")), ".polarys")
	if err != nil {
		panic(err)
	}

	config := params.DefaultConfig
	chainParams := params.Polarys

	validators := make([]common.Address, 0)

	validators = append(validators, addr)

	engine := pow.InitConsensus(chainParams.PowEngine.Epoch, chainParams.PowEngine.Difficulty, chainParams.PowEngine.Delay, chainParams.ChainID, validators)

	blockchain, err := core.InitBlockchain(db, config, chainParams, engine, nil, logger)
	if err != nil {
		panic(err)
	}

	worker := miner.NewWorker(m, engine, blockchain, chainParams, logger)

	go worker.Run()

	select {}
}
