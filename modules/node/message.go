package node

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
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
	DIFF
	PEER_INFO
)

type Message struct {
	Type      Type   `json:"type"`
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
}

func NewMessage(t Type, d []byte, pubKey pec256.PubKey, aesKey []byte) (*Message, error) {
	buf := make([]byte, len(d)+32+16+8) // Data length + pubkey + nonce + timestamp
	copy(buf, d)
	copy(buf[len(d):], pubKey[:])

	nonce := utils.SecureNonce(16)
	now := uint64(time.Now().Unix())

	copy(buf[len(d)+32:], nonce)
	tBytes := common.Uint64ToBytes(now)
	copy(buf[len(d)+32+16:], tBytes)

	encrypted, err := encryptPayload(aesKey, buf)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: t,
		Data: encrypted,
	}, nil
}

func (m *Message) DecryptData(aesKey []byte) (*Message, error) {
	aux := copyMessage(m)
	d, err := decryptPayload(aesKey, m.Data)
	if err != nil {
		return nil, err
	}
	aux.Data = d
	m = aux
	return m, nil
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

func encryptPayload(aesKey, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	// Enviamos nonce|ct para descifrar luego
	return append(nonce, ct...), nil
}

// Descifra tras verificar la firma
func decryptPayload(aesKey, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("payload demasiado corto")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}
