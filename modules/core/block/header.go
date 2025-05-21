package block

import (
	"encoding/json"

	"github.com/polarysfoundation/polarys-chain/modules/common"
)

var (
	uint64Size  = 8
	hashSize    = 32
	addressSize = 15
)

type Header struct {
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

func (h *Header) CalculateSize() uint64 {
	size := uint64(0)

	// Adding 9 fields with uint64 value
	size += uint64(9 * uint64Size)

	// Adding prev(hash)
	size += uint64(hashSize)

	// Adding address size
	size += uint64(addressSize)

	// Adding validator proof size
	size += calcSliceSize(h.ValidatorProof)

	// Adding consensus proof size
	size += calcSliceSize(h.ConsensusProof)

	// Adding signature size
	size += calcSliceSize(h.Signature)

	// Adding data size
	size += calcSliceSize(h.Data)

	h.Size = size

	return size
}

func calcSliceSize(b []byte) uint64 {
	return uint64(len(b))
}

func (h *Header) Deserialize(data []byte) error {
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
		Size            uint64         `json:"size"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	h.Height = temp.Height
	h.Prev = temp.Prev
	h.Timestamp = temp.Timestamp
	h.Nonce = temp.Nonce
	h.GasTarget = temp.GasTarget
	h.GasTip = temp.GasTip
	h.GasUsed = temp.GasUsed
	h.Difficulty = temp.Difficulty
	h.TotalDifficulty = temp.TotalDifficulty
	h.Data = temp.Data
	h.ValidatorProof = temp.ValidatorProof
	h.ConsensusProof = temp.ConsensusProof
	h.Validator = temp.Validator
	h.Size = temp.Size

	return nil
}

func (h *Header) Serialize() ([]byte, error) {
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
		Size            uint64         `json:"size"`
	}{
		Height:          h.Height,
		Prev:            h.Prev,
		Timestamp:       h.Timestamp,
		Nonce:           h.Nonce,
		GasTarget:       h.GasTarget,
		GasTip:          h.GasTip,
		GasUsed:         h.GasUsed,
		Difficulty:      h.Difficulty,
		TotalDifficulty: h.TotalDifficulty,
		Data:            h.Data,
		ValidatorProof:  h.ValidatorProof,
		ConsensusProof:  h.ConsensusProof,
		Validator:       h.Validator,
		Size:            h.Size,
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil
}
