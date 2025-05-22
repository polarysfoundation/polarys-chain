package core

import (
	"encoding/json"
	"time"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/block"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

var (
	zeroHash    = common.Hash([32]byte{})
	zeroAddress = common.Address([15]byte{})
)

type GenesisBlock struct {
	Height          uint64         `json:"height"`
	Prev            common.Hash    `json:"prev"`
	Timestamp       uint64         `json:"timestamp"`
	Nonce           uint64         `json:"nonce"`
	GasTarget       uint64         `json:"gas_target"`
	GasTip          uint64         `json:"gas_tip"`
	GasUsed         uint64         `json:"gas_used"`
	Difficulty      uint64         `json:"difficulty"`
	TotalDifficulty uint64         `json:"total_difficulty"`
	Data            []byte         `json:"data"`
	ValidatorProof  []byte         `json:"validator_proof"`
	ConsensusProof  []byte         `json:"consensus_proof"`
	Signature       []byte         `json:"signature"`
	Validator       common.Address `json:"validator"`
	Size            uint64         `json:"size"`
}

func InitGenesisBlock(db *polarysdb.Database, dfChain bool, genesis *GenesisBlock, consensusProof []byte) (*GenesisBlock, error) {
	err := checkMetrics(db)
	if err != nil {
		return nil, err
	}

	if hasGenesisBlock(db) {
		return getGenesisBlock(db)
	}

	if dfChain && genesis == nil {
		gb := defaultBlock(consensusProof)

		blk, err := gb.ToBlock()
		if err != nil {
			return nil, err
		}

		err = saveBlock(db, blk)
		if err != nil {
			return nil, err
		}
		return gb, nil
	} else {
		blk, err := genesis.ToBlock()
		if err != nil {
			return nil, err
		}

		err = saveBlock(db, blk)
		if err != nil {
			return nil, err
		}

		return genesis, nil
	}
}

func (g *GenesisBlock) Hash() common.Hash {
	data, err := g.Serialize()
	if err != nil {
		panic(err)
	}

	return common.BytesToHash(data)
}

func defaultBlock(consensusProof []byte) *GenesisBlock {
	gb := &GenesisBlock{
		Height:          0,
		Prev:            zeroHash,
		Timestamp:       uint64(time.Now().Unix()),
		Nonce:           0,
		GasTarget:       0,
		GasTip:          0,
		GasUsed:         0,
		Difficulty:      0,
		TotalDifficulty: 0,
		Data:            []byte{},
		ValidatorProof:  []byte{},
		ConsensusProof:  consensusProof,
		Validator:       zeroAddress,
	}

	return gb
}

func (g *GenesisBlock) Deserialize(data []byte) error {
	temp := struct {
		Height          uint64         `json:"height"`
		Prev            common.Hash    `json:"prev"`
		Timestamp       uint64         `json:"timestamp"`
		Nonce           uint64         `json:"nonce"`
		GasTarget       uint64         `json:"gas_target"`
		GasTip          uint64         `json:"gas_tip"`
		GasUsed         uint64         `json:"gas_used"`
		Difficulty      uint64         `json:"difficulty"`
		TotalDifficulty uint64         `json:"total_difficulty"`
		Data            []byte         `json:"data"`
		ValidatorProof  []byte         `json:"validator_proof"`
		ConsensusProof  []byte         `json:"consensus_proof"`
		Validator       common.Address `json:"validator"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	g.Height = temp.Height
	g.Prev = temp.Prev
	g.Timestamp = temp.Timestamp
	g.Nonce = temp.Nonce
	g.GasTarget = temp.GasTarget
	g.GasTip = temp.GasTip
	g.GasUsed = temp.GasUsed
	g.Difficulty = temp.Difficulty
	g.TotalDifficulty = temp.TotalDifficulty
	g.Data = temp.Data
	g.ValidatorProof = temp.ValidatorProof
	g.ConsensusProof = temp.ConsensusProof
	g.Validator = temp.Validator

	return nil
}

func (g *GenesisBlock) ToBlock() (*block.Block, error) {
	if g != nil {
		return block.NewBlock(block.Header{
			Height:          g.Height,
			Prev:            g.Prev,
			Timestamp:       g.Timestamp,
			Nonce:           g.Nonce,
			GasTarget:       g.GasTarget,
			GasTip:          g.GasTip,
			GasUsed:         g.GasUsed,
			Difficulty:      g.Difficulty,
			TotalDifficulty: g.TotalDifficulty,
			Data:            g.Data,
			ValidatorProof:  g.ValidatorProof,
			ConsensusProof:  g.ConsensusProof,
			Signature:       g.Signature,
			Validator:       g.Validator,
			Size:            g.Size,
		}, nil), nil
	}

	return nil, ErrBlockNotInitialized
}

func (g *GenesisBlock) Serialize() ([]byte, error) {
	temp := struct {
		Height          uint64         `json:"height"`
		Prev            common.Hash    `json:"prev"`
		Timestamp       uint64         `json:"timestamp"`
		Nonce           uint64         `json:"nonce"`
		GasTarget       uint64         `json:"gas_target"`
		GasTip          uint64         `json:"gas_tip"`
		GasUsed         uint64         `json:"gas_used"`
		Difficulty      uint64         `json:"difficulty"`
		TotalDifficulty uint64         `json:"total_difficulty"`
		Data            []byte         `json:"data"`
		ValidatorProof  []byte         `json:"validator_proof"`
		ConsensusProof  []byte         `json:"consensus_proof"`
		Validator       common.Address `json:"validator"`
	}{}

	temp.Height = g.Height
	temp.Prev = g.Prev
	temp.Timestamp = g.Timestamp
	temp.Nonce = g.Nonce
	temp.GasTarget = g.GasTarget
	temp.GasTip = g.GasTip
	temp.GasUsed = g.GasUsed
	temp.Difficulty = g.Difficulty
	temp.TotalDifficulty = g.TotalDifficulty
	temp.Data = g.Data
	temp.ValidatorProof = g.ValidatorProof
	temp.ConsensusProof = g.ConsensusProof
	temp.Validator = g.Validator

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func checkMetrics(db *polarysdb.Database) error {
	if !db.Exist(metricCurrent) {
		err := db.Create(metricCurrent)
		if err != nil {
			return err
		}
	}

	if !db.Exist(metricByNumber) {
		err := db.Create(metricByNumber)
		if err != nil {
			return err
		}
	}

	if !db.Exist(metricByHash) {
		err := db.Create(metricByHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func hasGenesisBlock(db *polarysdb.Database) bool {
	_, ok := db.Read(metricCurrent, "0")
	return ok
}

func getGenesisBlock(db *polarysdb.Database) (*GenesisBlock, error) {
	data, ok := db.Read(metricByNumber, "0")
	if !ok {
		return nil, ErrBlockNotFound
	}

	var g *block.Block

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &g)
	if err != nil {
		return nil, err
	}

	gb := &GenesisBlock{
		Height:          g.Height(),
		Prev:            g.Prev(),
		Timestamp:       g.Timestamp(),
		Nonce:           g.Nonce(),
		GasTarget:       g.GasTarget(),
		GasTip:          g.GasTip(),
		GasUsed:         g.GasUsed(),
		Difficulty:      g.Difficulty(),
		TotalDifficulty: g.TotalDifficulty(),
		Data:            g.Data(),
		ValidatorProof:  g.ValidatorProof(),
		ConsensusProof:  g.ConsensusProof(),
		Signature:       g.Signature(),
		Validator:       g.Validator(),
		Size:            g.Size(),
	}

	return gb, nil
}
