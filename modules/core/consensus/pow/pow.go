package pow

import (
	"errors"
	"log"
	"math/big"

	"slices"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
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

func InitConsensus(epoch, difficulty, delay, chainID uint64, validators []common.Address) *Consensus {
	buff := common.Decode("PowEngine")
	protocolHash := crypto.Pm256(buff)

	return &Consensus{
		epoch:        epoch,
		difficulty:   difficulty,
		delay:        delay,
		validators:   validators,
		protocolHash: common.BytesToHash(protocolHash),
		chainID:      chainID,
	}
}

func (c *Consensus) Validator() common.Address {
	return c.currentValidator
}

func (c *Consensus) ProtocolHash() common.Hash {
	return c.protocolHash
}

func (c *Consensus) ConsensusProof(crrBlockNumber uint64) ([]byte, error) {
	if crrBlockNumber == 0 {
		return nil, errors.New("invalid block number")
	}

	consensusProof := make([]byte, 64)
	buff := make([]byte, 32)
	copy(buff[:8], common.Uint64ToBytes(c.chainID)) // Fixed copy direction
	copy(buff[8:16], common.Uint64ToBytes(crrBlockNumber))
	copy(buff[16:24], common.Uint64ToBytes(c.epoch))
	copy(buff[24:32], common.Uint64ToBytes(uint64(len(c.validators))))
	copy(consensusProof[:32], buff)
	copy(consensusProof[32:], c.protocolHash.Bytes())

	return consensusProof, nil
}

func (c *Consensus) ValidatorProof() ([]byte, error) {
	validatorProof := make([]byte, 64)

	copy(validatorProof[:8], common.Uint64ToBytes(c.chainID))
	copy(validatorProof[8:16], common.Uint64ToBytes(c.epoch)) // Fixed offset
	copy(validatorProof[16:31], c.currentValidator.Bytes())   // Address is 32 bytes
	copy(validatorProof[32:64], c.protocolHash.Bytes())       // Fixed offset

	return validatorProof, nil
}

func (c *Consensus) SealBlock(block *block.Block) (*block.Block, error) {
	if block == nil {
		return nil, ErrNilBlock
	}

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

	return block, nil
}

func (c *Consensus) VerifyBlock(chain consensus.Chain, block *block.Block) (bool, error) {
	if block == nil {
		return false, ErrNilBlock
	}

	log.Println(block.Height())

	prevBlock, err := chain.GetBlockByHeight(block.Height() - 1)
	if err != nil {
		return false, err
	}

	if prevBlock == nil {
		return false, ErrNilPreviousBlock
	}

	if prevBlock.Hash() != block.Prev() {
		return false, ErrInvalidBlockHash
	}

	if ok, err := c.verifyConsensusProof(block, prevBlock); err != nil || !ok {
		return false, err
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

	if block.Height() <= latestBlock.Height() {
		return false, ErrInvalidBlockHeight
	}

	if block.Prev() != latestBlock.Hash() {
		return false, ErrInvalidBlockHash
	}

	tmpBlk, err := chain.GetBlockByHeight(block.Height())
	if err == nil && tmpBlk != nil {
		return false, ErrDuplicatedBlock
	}

	if !block.SealHash().IsValid() {
		return false, ErrInvalidSealHash
	}

	return true, nil
}

func (c *Consensus) VerifyChain(chain consensus.Chain) (bool, error) {
	latestBlock, err := chain.GetLatestBlock()
	if err != nil {
		return false, err
	}

	if latestBlock.Height() == 0 {
		return true, nil
	}

	for i := uint64(2); i <= latestBlock.Height(); i++ { // Start from 1 to avoid underflow
		currentBlock, err := chain.GetBlockByHeight(i)
		if err != nil {
			return false, err
		}

		prevBlock, err := chain.GetBlockByHeight(i - 1)
		if err != nil {
			return false, err
		}

		if !common.Equal(currentBlock.Hash().Bytes(), currentBlock.CalcHash().Bytes()) {
			return false, ErrInvalidBlockHash
		}

		if !common.Equal(prevBlock.Hash().Bytes(), currentBlock.Prev().Bytes()) {
			return false, ErrInvalidBlockHash
		}

		if ok, err := c.DifficultyValidator(currentBlock, prevBlock); err != nil || !ok {
			return false, ErrInvalidDifficulty
		}
	}

	return true, nil
}

func (c *Consensus) verifyConsensusProof(block *block.Block, prevBlock *block.Block) (bool, error) {
	if block == nil || prevBlock == nil {
		return false, ErrNilBlock
	}

	consensusProof := block.ConsensusProof()
	if len(consensusProof) != 64 {
		return false, ErrInvalidConsensusProof
	}

	chainID := common.BytesToUint64(consensusProof[:8])
	crrBlockNumber := common.BytesToUint64(consensusProof[8:16])
	epoch := common.BytesToUint64(consensusProof[16:24])
	validatorCount := common.BytesToUint64(consensusProof[24:32])
	protocolHash := common.BytesToHash(consensusProof[32:])

	if crrBlockNumber == prevBlock.Height()+1 {
		return false, ErrInvalidBlockHeight
	}

	if chainID != c.chainID {
		return false, ErrInvalidChainID
	}

	if len(c.validators) != int(validatorCount) {
		return false, ErrInvalidValidatorCount
	}

	if epoch != c.epoch {
		return false, ErrInvalidEpoch
	}

	if protocolHash != c.protocolHash {
		return false, ErrInvalidProtocolHash
	}

	if !c.ValidatorExists(block.Validator()) {
		return false, ErrInvalidValidator
	}

	return true, nil
}

func (c *Consensus) verifyValidatorProof(block *block.Block) bool {
	if block == nil {
		return false
	}

	validatorProof := block.ValidatorProof()
	if len(validatorProof) != 64 {
		return false
	}

	chainID := common.BytesToUint64(validatorProof[:8])
	epoch := common.BytesToUint64(validatorProof[8:16])
	validator := common.BytesToAddress(validatorProof[16:31]) // Address is 32 bytes
	protocolHash := common.BytesToHash(validatorProof[32:])

	log.Println(validator.String())
	log.Println(block.Validator().String())

	if chainID != c.chainID {
		return false
	}

	if epoch != c.epoch {
		return false
	}

	if protocolHash != c.protocolHash {
		return false
	}

	return validator == block.Validator()
}

func (c *Consensus) ValidatorExists(address common.Address) bool {
	return slices.Contains(c.validators, address)
}

func (c *Consensus) AdjustDifficulty(block *block.Block, prevBlock *block.Block) uint64 {
	// Siempre recalculamos en cada bloque, para que builder y validator
	// hablen el mismo idioma:
	if block == nil || prevBlock == nil || block.Height() == 0 {
		// bloque 0 o nil: sin cambio
		return c.difficulty
	}

	newDifficulty := c.calcDifficulty(block, prevBlock)
	// guardamos el valor, respetando rangos
	bounded := max(min(newDifficulty, MaxDifficulty), MinDifficulty)
	c.difficulty = bounded
	c.lastAdjustment = block.Height()
	c.lastDifficulty = bounded
	return bounded
}

func (c *Consensus) DifficultyValidator(block *block.Block, prevBlock *block.Block) (bool, error) {
	if block == nil || prevBlock == nil {
		return false, ErrNilBlock
	}

	if block.Height() == 0 {
		return true, nil
	}

	expectedDifficulty := c.calcDifficulty(block, prevBlock)
	blockDifficulty := block.Difficulty()

	// Allow a small margin for difficulty drift
	const margin = 0.1
	minDiff := uint64(float64(expectedDifficulty) * (1 - margin))
	maxDiff := uint64(float64(expectedDifficulty) * (1 + margin))

	return blockDifficulty >= minDiff && blockDifficulty <= maxDiff, nil
}

func (c *Consensus) SelectValidator() common.Address {
	if len(c.validators) == 0 {
		return common.Address{}
	}

	nextVal := c.validators[0]
	c.currentValidator = nextVal
	c.latestValidator = nextVal
	return nextVal
}

func (c *Consensus) calcDifficulty(block *block.Block, prevBlock *block.Block) uint64 {
	if block == nil || prevBlock == nil {
		return c.difficulty
	}

	prevDifficulty := prevBlock.Difficulty()
	if block.Height() == 0 {
		return prevDifficulty
	}

	gasUsed := block.GasUsed()
	gasTarget := block.GasTarget()
	var gasRatio float64
	if gasUsed == 0 {
		gasRatio = 1.0
	} else {
		gasRatio = float64(gasTarget) / float64(gasUsed)
	}

	blockTime := block.Timestamp()
	prevBlockTime := prevBlock.Timestamp()
	timeDelta := blockTime - prevBlockTime
	if timeDelta == 0 {
		timeDelta = 1
	}

	targetBlockTime := float64(c.delay)
	timeFactor := targetBlockTime / float64(timeDelta)

	// Calculate adjustment factor
	adjustment := 0.7*timeFactor + 0.3*gasRatio

	// Apply bounds to adjustment
	if adjustment > 1.2 {
		adjustment = 1.2
	} else if adjustment < 0.8 {
		adjustment = 0.8
	}

	// Calculate new difficulty
	newDifficulty := float64(prevDifficulty) * adjustment

	// Apply bounds to difficulty
	if newDifficulty > float64(MaxDifficulty) {
		newDifficulty = float64(MaxDifficulty)
	} else if newDifficulty < float64(MinDifficulty) {
		newDifficulty = float64(MinDifficulty)
	}

	return uint64(newDifficulty)
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
