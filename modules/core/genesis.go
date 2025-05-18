package core

import (
	"encoding/json"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	polarysdb "github.com/polarysfoundation/polarys_db"
)

var (
	metricByNumber = "block/number/"
	metricByHash   = "block/hash/"
	metricCurrent  = "block/current/"
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
}

func InitGenesisBlock(db *polarysdb.Database, dfChain bool, genesis *GenesisBlock) (*GenesisBlock, error) {
	err := checkMetrics(db)
	if err != nil {
		return nil, err
	}

	if hasGenesisBlock(db) {
		return getGenesisBlock(db)
	}

	if dfChain && genesis == nil {
		gb := defaultBlock()
		err = saveGenesisBlock(db, gb)
		if err != nil {
			return nil, err
		}
		return gb, nil
	} else {
		err = saveGenesisBlock(db, genesis)
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

func defaultBlock() *GenesisBlock {
	gb := &GenesisBlock{
		Height:          0,
		Prev:            zeroHash,
		Timestamp:       0,
		Nonce:           0,
		GasTarget:       0,
		GasTip:          0,
		GasUsed:         0,
		Difficulty:      0,
		TotalDifficulty: 0,
		Data:            []byte{},
		ValidatorProof:  []byte{},
		ConsensusProof:  []byte{},
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

func saveGenesisBlock(db *polarysdb.Database, g *GenesisBlock) error {
	err := db.Write(metricByNumber, "0", g)
	if err != nil {
		return err
	}
	err = db.Write(metricByHash, g.Hash().String(), g)
	if err != nil {
		return err
	}

	err = db.Write(metricCurrent, "0", g)
	if err != nil {
		return err
	}

	return nil
}

func getGenesisBlock(db *polarysdb.Database) (*GenesisBlock, error) {
	data, ok := db.Read(metricByNumber, "0")
	if !ok {
		return nil, ErrBlockNotFound
	}

	var g GenesisBlock

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = g.Deserialize(b)
	if err != nil {
		return nil, err
	}

	return &g, nil
}
