package pow

import (
	"math/big"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	"github.com/polarysfoundation/polarys-chain/modules/core/consensus"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

var (
	MaxDifficulty = uint64(0xFFFFFFFFFFFFFFFF)
	MaxNonce      = uint64(0xFFFFFFFFFFFFFFFF)
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
	if !c.verifyConsensusProof(block) {
		return false, ErrInvalidConsensusProof
	}

	if !c.verifyValidatorProof(block) {
		return false, ErrInvalidValidatorProof
	}

	if !c.ValidatorExists(block.Validator()) {
		return false, ErrInvalidValidator
	}

	if !c.DifficultyValidator(block) {
		return false, ErrInvalidDifficulty
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

func (c *Consensus) verifyConsensusProof(block *block.Block) bool {
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

	if !c.DifficultyValidator(block) {
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
	for _, validator := range c.validators {
		if validator == address {
			return true
		}
	}
	return false
}

func (c *Consensus) DifficultyValidator(block *block.Block) bool {
	return true
}

// This function work just with a single validator for PoW protocol
func (c *Consensus) SelectValidator() common.Address {
	nextVal := c.validators[0]

	c.currentValidator = nextVal
	c.latestValidator = nextVal
	return nextVal
}

func (c *Consensus) calcDifficulty() {

}
