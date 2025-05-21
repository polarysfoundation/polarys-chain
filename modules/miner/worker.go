package miner

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/core"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/params"
)

type Worker struct {
	miner      *Miner
	engine     consensus.Engine
	blockchain *core.Blockchain
	config     *params.ChainParams
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewWorker(miner *Miner, engine consensus.Engine, blockchain *core.Blockchain, config *params.ChainParams) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		miner:      miner,
		engine:     engine,
		blockchain: blockchain,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (w *Worker) Run() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(time.Duration(w.config.PowEngine.Delay) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				w.tryProduceBlock()
			}
		}
	}()
}

func (w *Worker) tryProduceBlock() {
	latest, err := w.blockchain.GetLatestBlock()
	if err != nil {
		log.Println("Failed to get latest block:", err)
		return
	}

	nonce := calcNewNonce(latest.Nonce())
	if nonce == 0 || nonce == latest.Nonce() || nonce == ^latest.Nonce() {
		return
	}

	selectedTxs, gasUsed, gasTip := w.selectTransactions()
	if len(selectedTxs) == 0 {
		return
	}

	validatorProof, err := w.engine.ValidatorProof()
	if err != nil {
		log.Println("Validator proof error:", err)
		return
	}

	consensusProof, err := w.engine.ConsensusProof(latest.Height())
	if err != nil {
		log.Println("Consensus proof error:", err)
		return
	}

	header := w.buildHeader(latest, nonce, gasUsed, gasTip, validatorProof, consensusProof)
	newBlock := block.NewBlock(header, selectedTxs)

	if err := newBlock.SignBlock(w.miner.privKey); err != nil {
		log.Println("Block signing failed:", err)
		return
	}

	newBlock, err = w.engine.SealBlock(newBlock)
	if err != nil {
		log.Println("Sealing failed:", err)
		return
	}

	if !w.engine.VerifyBlock(newBlock) {
		log.Println("Block verification failed")
		return
	}

	if err := w.blockchain.AddBlock(newBlock); err != nil {
		log.Println("Failed to add block:", err)
		return
	}

	if _, err := w.engine.VerifyChain(w.blockchain); err != nil {
		log.Println("Chain verification failed:", err)
	}
}

func (w *Worker) selectTransactions() ([]transaction.Transaction, uint64, uint64) {
	txs := w.blockchain.GetTransactions()
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
	return selected, gasUsed, gasTip
}

func (w *Worker) buildHeader(prev *block.Block, nonce, gasUsed, gasTip uint64, valProof, consProof []byte) block.Header {
	header := block.Header{
		Height:         prev.Height() + 1,
		Prev:           prev.Hash(),
		Timestamp:      uint64(time.Now().Unix()),
		Nonce:          nonce,
		GasTarget:      w.blockchain.GasTarget(),
		GasTip:         gasTip,
		GasUsed:        gasUsed,
		Difficulty:     w.blockchain.Difficulty(),
		Data:           []byte{},
		Validator:      w.miner.address,
		ValidatorProof: valProof,
		ConsensusProof: consProof,
	}
	header.CalculateSize()
	return header
}

func calcNewNonce(prevNonce uint64) uint64 {
	var randBytes [8]byte
	_, err := rand.Read(randBytes[:])
	if err != nil {
		return ^prevNonce ^ uint64(time.Now().UnixNano())
	}
	randomPart := binary.LittleEndian.Uint64(randBytes[:])
	timePart := uint64(time.Now().UnixNano())
	nonce := prevNonce ^ randomPart ^ timePart

	offset, err := rand.Int(rand.Reader, big.NewInt(1<<16))
	if err == nil {
		nonce ^= offset.Uint64()
	}
	return nonce
}

func (w *Worker) Stop() {
	w.cancel()
	w.wg.Wait()
}
