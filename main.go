// --- cmd/node/main.go ---

package main

import (
	"os"
	"os/signal"
	"syscall"

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
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// Inicialización habitual...
	accounts := accounts.InitAccounts(logger)
	addr, _ := accounts.NewAccount([]byte("test"))
	db, _ := polarysdb.Init(polarysdb.GenerateKeyFromBytes([]byte("test")), ".polarys")

	config := params.DefaultConfig
	chainParams := params.Polarys
	engine := pow.InitConsensus(chainParams.PowEngine.Epoch, chainParams.PowEngine.Difficulty, chainParams.PowEngine.Delay, chainParams.ChainID, []common.Address{addr})

	blockchain, _ := core.InitBlockchain(db, config, chainParams, engine, nil, logger)
	engine.SelectValidator()

	// Arrancamos los loops de blockchain y el worker de minería
	blockchain.Start()
	worker := miner.NewWorker(miner.NewMiner(addr, accounts), engine, blockchain, chainParams, logger)
	worker.Run()

	// Capturamos señal de interrupción para apagar en limpio
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	logger.Info("Shutting down node...")

	// Paramos componentes en orden
	worker.Stop()
	blockchain.Stop()

	logger.Info("Node terminated")
}
