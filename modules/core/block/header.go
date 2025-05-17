package block

import (
	"encoding/json"

	"github.com/polarysfoundation/polarys-chain/modules/common"
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
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return b, nil
}
