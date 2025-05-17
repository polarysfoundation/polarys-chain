package block

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/core/transaction"
)

func newTestHeader() Header {
	prevHash := common.BytesToHash([]byte("previous_block_hash_1234567890"))
	validatorAddr := common.BytesToAddress([]byte("validator_address_0x123"))
	return Header{
		Height:          10,
		Prev:            prevHash,
		Timestamp:       1678886400, // Example timestamp
		Nonce:           12345,
		GasTarget:       5000000,
		GasTip:          2,
		GasUsed:         4500000,
		Difficulty:      1000,
		TotalDifficulty: 200000,
		Data:            []byte("block specific data"),
		ValidatorProof:  []byte("validator proof data"),
		ConsensusProof:  []byte("consensus proof data"),
		Signature:       []byte("block signature data"),
		Validator:       validatorAddr,
	}
}

func TestNewBlock(t *testing.T) {
	header := newTestHeader()
	txs := []transaction.Transaction{
		{}, // Using zero-value transaction.Transaction for simplicity
		{},
	}

	block := NewBlock(header, txs)

	if !reflect.DeepEqual(block.header, header) {
		t.Errorf("NewBlock() header = %v, want %v", block.header, header)
	}
	if !reflect.DeepEqual(block.transactions, txs) {
		t.Errorf("NewBlock() transactions = %v, want %v", block.transactions, txs)
	}

	// Check that hash is not initially set (is invalid)
	if block.hash.IsValid() {
		t.Errorf("NewBlock() initial hash should be invalid, but got %v", block.hash)
	}
}

func TestBlock_Accessors(t *testing.T) {
	header := newTestHeader()
	txs := []transaction.Transaction{{}, {}}
	block := NewBlock(header, txs)

	if height := block.Height(); height != header.Height {
		t.Errorf("block.Height() = %v, want %v", height, header.Height)
	}
	if prev := block.Prev(); !reflect.DeepEqual(prev, header.Prev) {
		t.Errorf("block.Prev() = %v, want %v", prev, header.Prev)
	}
	if timestamp := block.Timestamp(); timestamp != header.Timestamp {
		t.Errorf("block.Timestamp() = %v, want %v", timestamp, header.Timestamp)
	}
	if nonce := block.Nonce(); nonce != header.Nonce {
		t.Errorf("block.Nonce() = %v, want %v", nonce, header.Nonce)
	}
	if gasTarget := block.GasTarget(); gasTarget != header.GasTarget {
		t.Errorf("block.GasTarget() = %v, want %v", gasTarget, header.GasTarget)
	}
	if gasTip := block.GasTip(); gasTip != header.GasTip {
		t.Errorf("block.GasTip() = %v, want %v", gasTip, header.GasTip)
	}
	if gasUsed := block.GasUsed(); gasUsed != header.GasUsed {
		t.Errorf("block.GasUsed() = %v, want %v", gasUsed, header.GasUsed)
	}
	if difficulty := block.Difficulty(); difficulty != header.Difficulty {
		t.Errorf("block.Difficulty() = %v, want %v", difficulty, header.Difficulty)
	}
	if totalDifficulty := block.TotalDifficulty(); totalDifficulty != header.TotalDifficulty {
		t.Errorf("block.TotalDifficulty() = %v, want %v", totalDifficulty, header.TotalDifficulty)
	}
	if data := block.Data(); !reflect.DeepEqual(data, header.Data) {
		t.Errorf("block.Data() = %v, want %v", data, header.Data)
	}
	if validatorProof := block.ValidatorProof(); !reflect.DeepEqual(validatorProof, header.ValidatorProof) {
		t.Errorf("block.ValidatorProof() = %v, want %v", validatorProof, header.ValidatorProof)
	}
	if consensusProof := block.ConsensusProof(); !reflect.DeepEqual(consensusProof, header.ConsensusProof) {
		t.Errorf("block.ConsensusProof() = %v, want %v", consensusProof, header.ConsensusProof)
	}
	if signature := block.Signature(); !reflect.DeepEqual(signature, header.Signature) {
		t.Errorf("block.Signature() = %v, want %v", signature, header.Signature)
	}
	if validator := block.Validator(); !reflect.DeepEqual(validator, header.Validator) {
		t.Errorf("block.Validator() = %v, want %v", validator, header.Validator)
	}
	if transactions := block.Transactions(); !reflect.DeepEqual(transactions, txs) {
		t.Errorf("block.Transactions() = %v, want %v", transactions, txs)
	}
}

func TestBlock_Hash(t *testing.T) {
	header := newTestHeader()
	block := NewBlock(header, nil)

	// 1. Test hash calculation and caching
	if block.hash.IsValid() {
		t.Fatalf("Initial block.hash should be invalid")
	}

	// Calculate expected hash
	headerBytes, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("Failed to marshal header for hash calculation: %v", err)
	}
	expectedHashBytes := pm256.Sum256(headerBytes)
	expectedHash := common.BytesToHash(expectedHashBytes[:])

	// First call to Hash()
	hash1 := block.Hash()
	if !reflect.DeepEqual(hash1, expectedHash) {
		t.Errorf("block.Hash() first call = %v, want %v", hash1, expectedHash)
	}
	if !block.hash.IsValid() {
		t.Errorf("block.hash should be valid after first Hash() call")
	}
	if !reflect.DeepEqual(block.hash, hash1) {
		t.Errorf("Internal block.hash not set correctly, got %v, want %v", block.hash, hash1)
	}

	// Second call to Hash() - should return cached hash
	hash2 := block.Hash()
	if !reflect.DeepEqual(hash2, hash1) {
		t.Errorf("block.Hash() second call = %v, want %v (cached)", hash2, hash1)
	}

	// Verify it's indeed the cached one by checking the internal field directly
	if !reflect.DeepEqual(block.hash, hash1) {
		t.Errorf("Internal block.hash %v changed or was not used, expected %v", block.hash, hash1)
	}
}

func TestBlock_Hash_Consistency(t *testing.T) {
	header1 := newTestHeader()
	// Ensure all fields are identical for header2
	header2 := Header{
		Height:          header1.Height,
		Prev:            header1.Prev,
		Timestamp:       header1.Timestamp,
		Nonce:           header1.Nonce,
		GasTarget:       header1.GasTarget,
		GasTip:          header1.GasTip,
		GasUsed:         header1.GasUsed,
		Difficulty:      header1.Difficulty,
		TotalDifficulty: header1.TotalDifficulty,
		Data:            append([]byte(nil), header1.Data...), // Create a copy
		ValidatorProof:  append([]byte(nil), header1.ValidatorProof...),
		ConsensusProof:  append([]byte(nil), header1.ConsensusProof...),
		Signature:       append([]byte(nil), header1.Signature...),
		Validator:       header1.Validator,
	}

	block1 := NewBlock(header1, nil)
	block2 := NewBlock(header2, nil)

	hash1 := block1.Hash()
	hash2 := block2.Hash()

	if !reflect.DeepEqual(hash1, hash2) {
		t.Errorf("Hashes for identical blocks differ: hash1 = %v, hash2 = %v", hash1, hash2)
		t.Logf("Header1: %+v", header1)
		t.Logf("Header2: %+v", header2)
	}
}

func TestBlock_Hash_Sensitivity(t *testing.T) {
	header1 := newTestHeader()
	block1 := NewBlock(header1, nil)
	hash1 := block1.Hash()

	// Create a slightly different header
	header2 := newTestHeader()
	header2.Nonce = header1.Nonce + 1 // Change one field

	block2 := NewBlock(header2, nil)
	hash2 := block2.Hash()

	if reflect.DeepEqual(hash1, hash2) {
		t.Errorf("Hashes for different blocks are identical: hash1 = %v, hash2 = %v", hash1, hash2)
		t.Logf("Header1: %+v", header1)
		t.Logf("Header2: %+v", header2)
	}
}