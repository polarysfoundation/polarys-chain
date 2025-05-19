package pow

import (
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	"slices"
)

var (
	MaxDifficulty      = uint64(0xFFFFFFFFFFFFFFFF)
	MaxNonce           = uint64(0xFFFFFFFFFFFFFFFF)
	DifficultyInterval = uint64(1000)
	MinDifficulty      = uint64(1)
	DifficultyDivisor  = uint64(10000)
)

var (
	BlockReward = big.NewInt(10000000000000000)
)

type Consensus struct {
	epoch            uint64
	difficulty       uint64
	delay            uint64
	validators       []common.Address
	protocolHash     common.Hash
	chainID          uint64
	currentValidator common.Address
	latestValidator  common.Address
	lastAdjustment   uint64
	lastDifficulty   uint64
}

func InitConsensus(epoch, difficulty, delay, chainID uint64) *Consensus {
	buff := common.Decode("PowEngine")
	protocolHash := crypto.Pm256(buff)

	return &Consensus{
		epoch:        epoch,
		difficulty:   difficulty,
		delay:        delay,
		validators:   make([]common.Address, 0),
		protocolHash: common.BytesToHash(protocolHash),
		chainID:      chainID,
	}
}

func (c *Consensus) ConsensusProof(crrBlockNumber uint64) ([]byte, error) {
	consensusProof := make([]byte, 64)
	buff := make([]byte, 32)
	copy(buff[8:], common.Uint64ToBytes(crrBlockNumber))
	copy(buff[:8], common.Uint64ToBytes(c.chainID))
	copy(buff[16:], common.Uint64ToBytes(uint64(len(c.validators))))
	copy(consensusProof[:32], buff)
	copy(consensusProof[32:], c.protocolHash.Bytes())

	return consensusProof, nil
}

func (c *Consensus) ValidatorProof() ([]byte, error) {
	validatorProof := make([]byte, 64)

	copy(validatorProof[:8], common.Uint64ToBytes(c.chainID))
	copy(validatorProof[:8], common.Uint64ToBytes(c.epoch))
	copy(validatorProof[16:], c.currentValidator.Bytes())
	copy(validatorProof[1:], common.Uint64ToBytes(uint64(len(c.validators))))
	copy(validatorProof[32:], c.protocolHash.Bytes())

	return validatorProof, nil
}

func (c *Consensus) SealBlock(block *block.Block) ([]byte, error) {
	consensusProof, err := c.ConsensusProof(block.Height())
	if err != nil {
		return nil, err
	}

	validatorProof, err := c.ValidatorProof()
	if err != nil {
		return nil, err
	}

	h1 := crypto.Pm256(consensusProof)
	h2 := crypto.Pm256(validatorProof)

	buff := make([]byte, 64)
	copy(buff[:32], h1)
	copy(buff[32:], h2)

	sealHash := crypto.Pm256(buff)
	block.Seal(common.BytesToHash(sealHash))

	return sealHash, nil
}

func (c *Consensus) VerifyBlock(chain consensus.Chain, block *block.Block) (bool, error) {
	prevBlock, err := chain.GetBlockByHeight(block.Height() - 1)
	if err != nil {
		return false, err
	}

	if prevBlock.Hash() != block.Prev() {
		return false, ErrInvalidBlockHash
	}

	if !c.verifyConsensusProof(block, prevBlock) {
		return false, ErrInvalidConsensusProof
	}

	if !c.verifyValidatorProof(block) {
		return false, ErrInvalidValidatorProof
	}

	if !c.ValidatorExists(block.Validator()) {
		return false, ErrInvalidValidator
	}

	latestBlock, err := chain.GetLatestBlock()
	if err != nil {
		return false, err
	}

	if block.Height() != latestBlock.Height()+1 {
		return false, ErrInvalidBlockHeight
	}

	if block.Prev() != latestBlock.Hash() {
		return false, ErrInvalidBlockHash
	}

	if block.Height()%c.epoch != 0 {
		return false, ErrInvalidEpoch
	}

	tmpBlk, err := chain.GetBlockByHeight(block.Height())
	if err != nil {
		return false, err
	}

	if tmpBlk.Height() == block.Height() {
		return false, ErrDuplicatedBlock
	}

	if !block.SealHash().IsValid() {
		return false, ErrInvalidSealHash
	}

	return true, nil
}

func (c *Consensus) verifyConsensusProof(block *block.Block, prevBlock *block.Block) bool {
	consensusProof := block.ConsensusProof()
	if len(consensusProof) != 64 {
		return false
	}
	chainID := common.BytesToUint64(consensusProof[:8])
	crrBlockNumber := common.BytesToUint64(consensusProof[8:16])
	epoch := common.BytesToUint64(consensusProof[16:24])
	validatorCount := common.BytesToUint64(consensusProof[24:32])
	protocolHash := common.BytesToHash(consensusProof[32:])
	if crrBlockNumber != block.Height() {
		return false
	}

	if chainID != c.chainID {
		return false
	}

	if len(c.validators) != int(validatorCount) {
		return false
	}

	if epoch != c.epoch {
		return false
	}

	if protocolHash != c.protocolHash {
		return false
	}

	if !c.ValidatorExists(block.Validator()) {
		return false
	}

	if !c.DifficultyValidator(block, prevBlock) {
		return false
	}

	return true

}

func (c *Consensus) verifyValidatorProof(block *block.Block) bool {
	validatorProof := block.ValidatorProof()
	if len(validatorProof) != 64 {
		return false
	}
	chainID := common.BytesToUint64(validatorProof[:8])
	epoch := common.BytesToUint64(validatorProof[8:16])
	validator := common.BytesToAddress(validatorProof[16:32])
	protocolHash := common.BytesToHash(validatorProof[32:])
	if chainID != c.chainID {
		return false
	}

	if epoch != c.epoch {
		return false
	}

	if protocolHash != c.protocolHash {
		return false
	}

	if validator != block.Validator() {
		return false
	}

	return true
}

func (c *Consensus) ValidatorExists(address common.Address) bool {
	return slices.Contains(c.validators, address)
}

func (c *Consensus) AdjustDifficulty(block *block.Block, prevBlock *block.Block) uint64 {
	if block.Height()%c.epoch != 0 {
		return c.difficulty
	}

	if block.Height() == 0 {
		c.difficulty = MinDifficulty
		return c.difficulty
	}

	newDifficulty := c.calcDifficulty(block, prevBlock)
	if newDifficulty != c.difficulty {
		c.difficulty = newDifficulty
		c.lastAdjustment = block.Height()
		c.lastDifficulty = newDifficulty
	}

	if c.difficulty < MinDifficulty {
		c.difficulty = MinDifficulty
	} else if c.difficulty > MaxDifficulty {
		c.difficulty = MaxDifficulty
	}

	return c.difficulty
}

// DifficultyValidator checks if the block's difficulty is valid based on expected difficulty.
// It compares the block's difficulty with the calculated difficulty using previous block data.
func (c *Consensus) DifficultyValidator(block *block.Block, prevBlock *block.Block) bool {
	// Genesis block: always valid
	if block.Height() == 0 {
		return true
	}

	// Get previous block (must be implemented in your chain)
	if prevBlock == nil {
		return false
	}

	expectedDifficulty := c.calcDifficulty(block, prevBlock)
	blockDifficulty := block.Difficulty()

	// Allow a small margin for difficulty drift (optional)
	const margin = 0 // set to 0 for strict equality, or small value for tolerance

	if blockDifficulty < expectedDifficulty-margin || blockDifficulty > expectedDifficulty+margin {
		return false
	}
	return true
}

// This function work just with a single validator for PoW protocol
func (c *Consensus) SelectValidator() common.Address {
	nextVal := c.validators[0]

	c.currentValidator = nextVal
	c.latestValidator = nextVal
	return nextVal
}

// calcDifficulty calculates the next block difficulty based on recent block data.
// It uses gas usage, block time, and previous difficulty to adjust the difficulty dynamically.
func (c *Consensus) calcDifficulty(block *block.Block, prevBlock *block.Block) uint64 {
	// Get previous difficulty
	prevDifficulty := c.difficulty
	if block.Height() == 0 {
		return prevDifficulty
	}

	// Get gas usage ratio (target/used)
	gasUsed := block.GasUsed()
	gasTarget := block.GasTarget()
	var gasRatio float64
	if gasUsed == 0 {
		gasRatio = 1.0 // Avoid division by zero, treat as optimal
	} else {
		gasRatio = float64(gasTarget) / float64(gasUsed)
	}

	// Get block time delta (current - previous)
	blockTime := block.Timestamp()
	prevBlockTime := prevBlock.Timestamp()
	timeDelta := blockTime - prevBlockTime
	if timeDelta == 0 {
		timeDelta = 1 // Avoid division by zero
	}

	// Calculate adjustment factor based on gas usage and block time
	// If blocks are too fast or gas is overused, increase difficulty
	// If blocks are too slow or gas is underused, decrease difficulty
	targetBlockTime := c.delay
	timeFactor := float64(targetBlockTime) / float64(timeDelta)

	// Weighted adjustment: 70% block time, 30% gas usage
	adjustment := max(min(0.7*timeFactor+0.3*gasRatio, 1.2), 0.8)

	// Calculate new difficulty
	newDifficulty := min(max(uint64(float64(prevDifficulty)*adjustment), MinDifficulty), MaxDifficulty)

	return newDifficulty
}
