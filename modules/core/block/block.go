package block

import (
	"encoding/json"
	"math/big"

	pec256 "github.com/polarysfoundation/pec-256"
	pm256 "github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
)

type Block struct {
	header       Header
	transactions []transaction.Transaction
	hash         common.Hash
}

func NewBlock(header Header, transactions []transaction.Transaction) *Block {
	return &Block{
		header:       header,
		transactions: transactions,
	}
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

	data, err := json.Marshal(b.header)
	if err != nil {
		panic(err)
	}

	hash := pm256.Sum256(data)

	b.hash = common.BytesToHash(hash[:])
	return b.hash
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
