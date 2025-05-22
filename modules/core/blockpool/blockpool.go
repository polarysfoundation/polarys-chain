package blockpool

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	"github.com/polarysfoundation/polarys-chain/modules/params"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

var (
	metric = "blockpool/"
)

type BlockPool struct {
	proposedBlocks  []*block.Block
	maxBlockSize    int64
	maxProposalSize int64
	slotHash        common.Hash // Slot hash always change each epoch creating a slot proposal each epoch per validators
	latestBlock     uint64
	chainID         uint64
	epoch           uint64

	engine consensus.Engine

	db   *polarysdb.Database
	lock sync.RWMutex
}

func NewBlockPool(engine consensus.Engine, db *polarysdb.Database, latestBlock uint64, config *params.Config, chainID uint64, epoch uint64) (*BlockPool, error) {
	if !db.Exist(metric) {
		err := db.Create(metric)
		if err != nil {
			return nil, err
		}
	}

	consensusProof, err := engine.ConsensusProof(latestBlock)
	if err != nil {
		return nil, err
	}

	slotHash := calcSlotHash(consensusProof, engine.Validator(), epoch, latestBlock, common.Hash{})

	poolBlock := &BlockPool{
		proposedBlocks:  make([]*block.Block, 0),
		maxBlockSize:    config.MaxBlockSize,
		maxProposalSize: config.MaxProposalSize,
		latestBlock:     latestBlock,
		engine:          engine,
		db:              db,
		chainID:         chainID,
		epoch:           epoch,
		slotHash:        slotHash,
	}

	return poolBlock, nil

}

func (pb *BlockPool) AddProposedBlock(block *block.Block) error {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	pb.proposedBlocks = append(pb.proposedBlocks, block)

	return nil
}

func (pb *BlockPool) ProcessProposedBlocks() (*block.Block, error) {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	validBlocks := make([]*block.Block, 0)
	for _, b := range pb.proposedBlocks {
		if pb.latestBlock == b.Height() {
			if b.Size() < uint64(pb.maxBlockSize) {
				validBlocks = append(validBlocks, b)
			}
		}
	}

	sort.Slice(validBlocks, func(i, j int) bool {
		return validBlocks[i].GasTarget() > validBlocks[j].GasTarget()
	})

	if len(validBlocks) == 0 {
		return nil, fmt.Errorf("no valid blocks found")
	}

	pb.latestBlock = validBlocks[0].Height()

	selectedBlock := validBlocks[0]
	selectedBlock.SetSlotHash(pb.slotHash)

	err := saveCurrentSlotHash(pb.db, selectedBlock.SlotHash())
	if err != nil {
		return nil, err
	}

	pb.proposedBlocks = make([]*block.Block, 0)

	return selectedBlock, nil
}

func (pb *BlockPool) SyncBlockPool(latestBlock uint64) error {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	pb.latestBlock = latestBlock

	consensusProof, err := pb.engine.ConsensusProof(latestBlock)
	if err != nil {
		return err
	}

	newSlotHash := calcSlotHash(consensusProof, pb.engine.Validator(), pb.epoch, latestBlock, pb.slotHash)

	pb.slotHash = newSlotHash

	return nil
}

func calcSlotHash(consensusProof []byte, validator common.Address, epoch, height uint64, parentHash common.Hash) common.Hash {
	buff := make([]byte, len(consensusProof)+len(validator)+8+8+32)
	copy(buff[:len(consensusProof)], consensusProof)
	copy(buff[len(consensusProof):], validator.Bytes())
	copy(buff[len(consensusProof)+len(validator):], common.Uint64ToBytes(epoch))
	copy(buff[len(consensusProof)+len(validator)+8:], common.Uint64ToBytes(height))
	copy(buff[len(consensusProof)+len(validator)+8+8:], parentHash.Bytes())

	h := crypto.Pm256(buff)

	return common.BytesToHash(h)
}

func getCurrentSlotHash(db *polarysdb.Database) (common.Hash, error) {
	crrIndex, err := getLatestSlotInded(db)
	if err != nil {
		return common.Hash{}, err
	}

	crrIndexStr := strconv.FormatUint(crrIndex, 10)

	data, ok := db.Read(metric, crrIndexStr)
	if !ok {
		return common.Hash{}, fmt.Errorf("slot hash not found")
	}

	h := data.(string)

	return common.CXIDToHash(h), nil
}

func saveCurrentSlotHash(db *polarysdb.Database, hash common.Hash) error {
	crrIndex, err := getLatestSlotInded(db)
	if err != nil {
		return err
	}

	crrIndexStr := strconv.FormatUint(crrIndex+1, 10)
	err = db.Write(metric, crrIndexStr, hash.String())
	if err != nil {
		return err
	}

	return nil
}

func getLatestSlotInded(db *polarysdb.Database) (uint64, error) {
	data, err := db.ReadBatch(metric)
	if err != nil {
		return 0, err
	}

	index := len(data)

	return uint64(index), nil
}
