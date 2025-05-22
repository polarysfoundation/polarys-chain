package miner

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/core"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	miner      *Miner
	engine     consensus.Engine
	blockchain *core.Blockchain
	config     *params.ChainParams
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	log        *logrus.Logger
}

func NewWorker(miner *Miner, engine consensus.Engine, blockchain *core.Blockchain, config *params.ChainParams, log *logrus.Logger) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	log.Info("Worker initialized")
	return &Worker{
		miner:      miner,
		engine:     engine,
		blockchain: blockchain,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
		log:        log,
	}
}

func (w *Worker) Run() {
	w.wg.Add(1)
	w.log.Info("Worker started")
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(time.Duration(w.config.PowEngine.Delay) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				w.log.Info("Worker stopped by context")
				return
			case <-ticker.C:
				w.tryProduceBlock()
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.log.Info("Stopping worker...")
	w.cancel()
	w.wg.Wait()
	w.log.Info("Worker stopped")
}

func (w *Worker) tryProduceBlock() {
	w.log.Info("Trying to produce new block...")
	latest, err := w.blockchain.GetLatestBlock()
	if err != nil {
		w.log.Error("Failed to get latest block", "err", err)
		return
	}

	nonce := calcNewNonce(latest.Nonce(), w.log)
	if nonce == 0 || nonce == latest.Nonce() || nonce == ^latest.Nonce() {
		w.log.Warn("Invalid nonce generated", "nonce", nonce)
		return
	}

	selectedTxs, gasUsed, gasTip := w.selectTransactions()

	validatorProof, err := w.engine.ValidatorProof()
	if err != nil {
		w.log.Error("Validator proof error", "err", err)
		return
	}

	consensusProof, err := w.engine.ConsensusProof(latest.Height())
	if err != nil {
		w.log.Error("Consensus proof error", "err", err)
		return
	}

	header := w.buildHeader(latest, nonce, gasUsed, gasTip, validatorProof, consensusProof)
	newBlock := block.NewBlock(header, selectedTxs)

	newBlock.CalcHash()

	newBlock, err = w.miner.SignBlock(newBlock, w.config.ChainID)
	if err != nil {
		w.log.Error("Block signing failed ", "err: ", err)
		return
	}

	if latest.Height() == newBlock.Height() {
		return
	}

	newBlock, err = w.engine.SealBlock(newBlock)
	if err != nil {
		w.log.Error("Block sealing failed ", "err: ", err)
		return
	}

	w.log.WithField("difficulty", newBlock.Difficulty()).Info("Block produced")

	if ok, err := w.engine.VerifyBlock(w.blockchain, newBlock); !ok || err != nil {
		w.log.Error("Block verification failed ", "ok: ", ok, " ", "err: ", err)
		return
	}

	if err := w.blockchain.AddBlock(newBlock); err != nil {
		w.log.Error("Failed to add block to blockchain ", "err: ", err)
		return
	}

	w.log.Info("Block produced and added ", "height: ", newBlock.Height(), " ", "hash: ", newBlock.Hash())

	if _, err := w.engine.VerifyChain(w.blockchain); err != nil {
		w.log.Error("Chain verification failed after block addition ", "err: ", err)
	}
}

func (w *Worker) selectTransactions() ([]transaction.Transaction, uint64, uint64) {
	txs := w.blockchain.GetTransactions()
	w.log.Debug("Selecting transactions ", "count: ", len(txs))
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GasPrice() > txs[j].GasPrice()
	})

	var (
		selected []transaction.Transaction
		gasUsed  uint64
		gasTip   uint64
		gasLimit = w.blockchain.GasTarget()
	)

	for _, tx := range txs {
		if gasUsed+tx.Gas() > gasLimit {
			continue
		}
		gasUsed += tx.Gas()
		gasTip += tx.GasTip()
		selected = append(selected, tx)
	}
	w.log.Debug("Transactions selected", "selected", len(selected), "gasUsed", gasUsed, "gasTip", gasTip)
	return selected, gasUsed, gasTip
}

func (w *Worker) buildHeader(prev *block.Block, nonce, gasUsed, gasTip uint64, valProof, consProof []byte) block.Header {
	w.log.Debug("Building block header", "prevHeight", prev.Height(), "nonce", nonce)

	// 1) Creamos un header provisional con la dificultad actual (la iremos ajustando)
	header := block.Header{
		Height:         prev.Height() + 1,
		Prev:           prev.Hash(),
		Timestamp:      uint64(time.Now().Unix()),
		Nonce:          nonce,
		GasTarget:      w.blockchain.GasTarget(),
		GasTip:         gasTip,
		GasUsed:        gasUsed,
		Difficulty:     prev.Difficulty(), // provisional
		Data:           []byte{},
		Validator:      w.miner.address,
		ValidatorProof: valProof,
		ConsensusProof: consProof,
	}

	w.log.Info("timestamp", header.Timestamp)

	// 2) Recalculamos la nueva dificultad según el motor de consenso
	//    para lo cual necesitamos un *block.Block provisional*:
	dummyBlock := block.NewBlock(header, nil)
	newDiff := w.engine.AdjustDifficulty(dummyBlock, prev)

	// 3) Asignamos la dificultad ajustada y recalculamos el tamaño
	header.Difficulty = newDiff
	header.CalculateSize()

	return header
}

func calcNewNonce(prevNonce uint64, log *logrus.Logger) uint64 {
	var randBytes [8]byte
	_, err := rand.Read(randBytes[:])
	if err != nil {
		log.Warn("Fallback random nonce due to read error", "err", err)
		return ^prevNonce ^ uint64(time.Now().UnixNano())
	}
	randomPart := binary.LittleEndian.Uint64(randBytes[:])
	timePart := uint64(time.Now().UnixNano())
	nonce := prevNonce ^ randomPart ^ timePart

	offset, err := rand.Int(rand.Reader, big.NewInt(1<<16))
	if err == nil {
		nonce ^= offset.Uint64()
	}
	log.Debug("Nonce generated ", "nonce: ", nonce)
	return nonce
}
