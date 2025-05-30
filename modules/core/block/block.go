package block

import (
	"encoding/json"

	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Block struct {
	header       Header
	transactions []transaction.Transaction
	hash         common.Hash
	sealHash     common.Hash
	slotHash     common.Hash
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

func (b *Block) MarshalJSON() ([]byte, error) {
	temp := struct {
		Header       Header      `json:"header"`
		Hash         common.Hash `json:"hash"`
		Transactions uint64      `json:"transactions"`
		SealHash     common.Hash `json:"seal_hash"`
		SlotHash     common.Hash `json:"slot_hash"`
	}{
		Header:       b.header,
		Hash:         b.hash,
		Transactions: uint64(len(b.transactions)),
		SealHash:     b.sealHash,
		SlotHash:     b.slotHash,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *Block) CalcHash() common.Hash {
	data, err := b.header.marshal()
	if err != nil {
		panic(err)
	}

	h := crypto.Pm256(data)
	b.hash = common.BytesToHash(h)
	return b.hash
}

func (b *Block) UnmarshalJSON(data []byte) error {
	temp := struct {
		Header       Header      `json:"header"`
		Hash         common.Hash `json:"hash"`
		Transactions uint64      `json:"transactions"`
		SealHash     common.Hash `json:"seal_hash"`
		SlotHash     common.Hash `json:"slot_hash"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	b.header = temp.Header
	b.hash = temp.Hash
	b.sealHash = temp.SealHash
	b.slotHash = temp.SlotHash
	b.transactions = make([]transaction.Transaction, temp.Transactions)

	return nil
}

func (b *Block) SetSlotHash(hash common.Hash) {
	b.slotHash = hash
}

func (b *Block) SlotHash() common.Hash {
	return b.slotHash
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

func (b *Block) SignBlock(signature []byte) (*Block, error) {
	auxBlock := copyBlock(b)
	auxBlock.header.Signature = signature

	return auxBlock, nil
}

func copyBlock(b *Block) *Block {
	transactions := make([]transaction.Transaction, len(b.transactions))
	copy(transactions, b.transactions)

	return &Block{
		header:       b.header,
		transactions: transactions,
		hash:         b.hash,
		sealHash:     b.sealHash,
		slotHash:     b.slotHash,
	}
}
