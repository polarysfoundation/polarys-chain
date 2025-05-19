package block

import (
	"encoding/json"
	"math/big"

	pec256 "github.com/polarysfoundation/pec-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Block struct {
	header       Header
	transactions []transaction.Transaction
	hash         common.Hash
	sealHash     common.Hash
}

func NewBlock(header Header, transactions []transaction.Transaction) *Block {

	size := header.CalculateSize()
	header.Size = size

	blk := &Block{
		header: header,
	}

	if len(transactions) > 0 {
		blk.transactions = transactions
	} else {
		blk.transactions = make([]transaction.Transaction, 0)
	}

	return blk
}

func (b *Block) Serialize() ([]byte, error) {
	temp := struct {
		Header       Header      `json:"header"`
		Hash         common.Hash `json:"hash"`
		Transactions uint64      `json:"transactions"`
		SealHash     common.Hash `json:"seal_hash"`
	}{
		Header:       b.header,
		Hash:         b.hash,
		Transactions: uint64(len(b.transactions)),
		SealHash:     b.sealHash,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *Block) CalcHash() common.Hash {
	data, err := b.header.Serialize()
	if err != nil {
		panic(err)
	}

	h := crypto.Pm256(data)
	b.hash = common.BytesToHash(h)
	return b.hash
}

func (b *Block) Deserialize(data []byte, transactions []transaction.Transaction) error {
	temp := struct {
		Header       Header      `json:"header"`
		Hash         common.Hash `json:"hash"`
		Transactions uint64      `json:"transactions"`
		SealHash     common.Hash `json:"seal_hash"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	temp.Header.Size = temp.Header.CalculateSize()

	b.header = temp.Header
	b.hash = temp.Hash
	b.sealHash = temp.SealHash

	b.transactions = make([]transaction.Transaction, temp.Transactions)
	copy(b.transactions, transactions)
	return nil
}

func (b *Block) AddTransaction(tx transaction.Transaction) {
	for _, t := range b.transactions {
		if t.Hash() == tx.Hash() {
			return
		}
	}

	b.transactions = append(b.transactions, tx)
}

func (b *Block) Timestamp() uint64 {
	return b.header.Timestamp
}

func (b *Block) Height() uint64 {
	return b.header.Height
}

func (b *Block) Prev() common.Hash {
	return b.header.Prev
}

func (b *Block) Transactions() []transaction.Transaction {
	return b.transactions
}

func (b *Block) Validator() common.Address {
	return b.header.Validator
}

func (b *Block) ValidatorProof() []byte {
	return b.header.ValidatorProof
}

func (b *Block) ConsensusProof() []byte {
	return b.header.ConsensusProof
}

func (b *Block) Signature() []byte {
	return b.header.Signature
}

func (b *Block) GasTarget() uint64 {
	return b.header.GasTarget
}

func (b *Block) GasTip() uint64 {
	return b.header.GasTip
}

func (b *Block) Size() uint64 {
	return b.header.Size
}

func (b *Block) GasUsed() uint64 {
	return b.header.GasUsed
}

func (b *Block) Difficulty() uint64 {
	return b.header.Difficulty
}

func (b *Block) TotalDifficulty() uint64 {
	return b.header.TotalDifficulty
}

func (b *Block) Data() []byte {
	return b.header.Data
}

func (b *Block) Nonce() uint64 {
	return b.header.Nonce
}

func (b *Block) Hash() common.Hash {
	if b.hash.IsValid() {
		return b.hash
	}

	return common.Hash{}
}

func (b *Block) SealHash() common.Hash {
	return b.sealHash
}

func (b *Block) Seal(seal common.Hash) {
	b.sealHash = seal
}

func (b *Block) SignBlock(priv pec256.PrivKey) error {
	data, err := b.header.Serialize()
	if err != nil {
		return err
	}

	h := crypto.Pm256(data)
	r, s, err := crypto.Sign(common.BytesToHash(h), priv)
	if err != nil {
		return err
	}

	signature := make([]byte, 64)
	copy(signature[:32], r.Bytes())
	copy(signature[32:], s.Bytes())

	b.header.Signature = signature

	return nil
}

func (b *Block) VerifyBlock(pub pec256.PubKey) (bool, error) {
	data, err := b.header.Serialize()
	if err != nil {
		return false, err
	}

	h := crypto.Pm256(data)

	r := new(big.Int).SetBytes(b.header.Signature[:32])
	s := new(big.Int).SetBytes(b.header.Signature[32:])

	return crypto.Verify(common.BytesToHash(h), r, s, pub)
}
