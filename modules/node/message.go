package node

import (
	"encoding/json"
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/utils"
)

type Type int

const (
	PING Type = iota
	PONG
	BLOCK
	HASH
	TRANSACTION
	ASK
)

type Message struct {
	Type      Type   `json:"type"`
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
}

func NewMessage(t Type, d []byte, pubKey pec256.PubKey) *Message {
	buf := make([]byte, len(d)+32+16+8) // Data length + pubkey + nonce + timestamp
	copy(buf, d)
	copy(buf[len(d):], pubKey[:])

	nonce := utils.SecureNonce(16)
	now := uint64(time.Now().Unix())

	copy(buf[len(d)+32:], nonce)
	tBytes := common.Uint64ToBytes(now)
	copy(buf[len(d)+32+16:], tBytes)

	return &Message{
		Type: t,
		Data: d,
	}
}

func (m *Message) DecodeNonce() ([]byte, error) {
	return m.Data[len(m.Data)-16 : 8], nil
}

func (m *Message) DecodeTimestamp() (uint64, error) {
	return common.BytesToUint64(m.Data[len(m.Data)-8:]), nil
}

func (m *Message) DecodePubKey() (pec256.PubKey, error) {
	var pubKey pec256.PubKey
	copy(pubKey[:], m.Data[len(m.Data)-32:])
	return pubKey, nil
}

func (m *Message) DecodeData() ([]byte, error) {
	return m.Data[:len(m.Data)-32], nil
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *Message) Bytes() []byte {
	b, _ := m.Marshal()
	return b
}

func (m *Message) SignMessage(s []byte) *Message {
	aux := copyMessage(m)

	aux.Signature = s
	m = aux
	return m
}

func copyMessage(m *Message) *Message {
	return &Message{
		Type:      m.Type,
		Data:      m.Data,
		Signature: m.Signature,
	}
}
