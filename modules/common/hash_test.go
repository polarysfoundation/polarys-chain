package common

import (
	"bytes"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestHash_SetBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Hash
	}{
		{
			name:  "nil bytes",
			input: nil,
			want:  Hash{}, // Expect all zeros, as copy(h[:], nil) does nothing to a zeroed Hash
		},
		{
			name:  "empty bytes",
			input: []byte{},
			want:  Hash{}, // Expect all zeros
		},
		{
			name:  "shorter than HashLen",
			input: []byte{0x01, 0x02, 0x03},
			want: func() Hash {
				var wantHash Hash // zero-initialized
				inputBytes := []byte{0x01, 0x02, 0x03}
				copy(wantHash[:len(inputBytes)], inputBytes) // Copies to the beginning
				return wantHash
			}(),
		},
		{
			name:  "exact HashLen",
			input: bytes.Repeat([]byte{0xAA}, HashLen),
			want: func() Hash {
				var wantHash Hash
				copy(wantHash[:], bytes.Repeat([]byte{0xAA}, HashLen))
				return wantHash
			}(),
		},
		{
			name:  "longer than HashLen",
			input: append(bytes.Repeat([]byte{0xBB}, 10), bytes.Repeat([]byte{0xCC}, HashLen)...), // Total HashLen + 10
			want: func() Hash {
				var wantHash Hash
				// SetBytes takes the last HashLen bytes from input
				expectedData := bytes.Repeat([]byte{0xCC}, HashLen)
				copy(wantHash[:], expectedData)
				return wantHash
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Hash
			got.SetBytes(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.SetBytes() with input %x = %x, want %x", tt.input, got, tt.want)
			}
		})
	}
}

func TestHash_StringAndCXID(t *testing.T) {
	hashData := bytes.Repeat([]byte{0xAB}, HashLen)
	var h Hash
	h.SetBytes(hashData)

	// Expected construction based on hash.go's cxid()
	// "1cx" + byte(HashByte) + 0x00 + hex(hashData) + trailing_nulls
	expectedPrefixAndByte := []byte{'1', 'c', 'x', byte(HashByte), 0x00} // 5 bytes
	hexEncodedData := make([]byte, HashLen*2)
	encode(hexEncodedData, hashData) // common.encode from bytes.go

	bufferLen := HashLen*3 + 2 // Current buffer allocation in hash.go's cxid()
	expectedBytes := make([]byte, bufferLen)
	copy(expectedBytes, expectedPrefixAndByte)
	copy(expectedBytes[len(expectedPrefixAndByte):], hexEncodedData)
	// The rest of expectedBytes (trailing part) is already 0x00 due to make.

	expectedStr := string(expectedBytes)

	if gotStr := h.String(); gotStr != expectedStr {
		t.Errorf("Hash.String() mismatch.\nGot : %q (len %d)\nWant: %q (len %d)", gotStr, len(gotStr), expectedStr, len(expectedStr))
		if !bytes.Equal([]byte(gotStr), []byte(expectedStr)) {
			t.Errorf("Hash.String() byte content mismatch.\nGot : %x\nWant: %x", []byte(gotStr), []byte(expectedStr))
		}
	}

	if gotCXID := h.CXID(); gotCXID != expectedStr {
		t.Errorf("Hash.CXID() mismatch.\nGot : %q (len %d)\nWant: %q (len %d)", gotCXID, len(gotCXID), expectedStr, len(expectedStr))
		if !bytes.Equal([]byte(gotCXID), []byte(expectedStr)) {
			t.Errorf("Hash.CXID() byte content mismatch.\nGot : %x\nWant: %x", []byte(gotCXID), []byte(expectedStr))
		}
	}
}

func TestHash_Hex(t *testing.T) {
	hashData := bytes.Repeat([]byte{0xBC}, HashLen)
	var h Hash
	h.SetBytes(hashData)

	// Expected construction based on hash.go's hex()
	// "0x" + hex(hashData)
	expectedPrefix := []byte{'0', 'x'}
	hexEncodedData := make([]byte, HashLen*2)
	encode(hexEncodedData, hashData) // common.encode from bytes.go

	expectedBufferLen := HashLen*2 + 2 // Current buffer allocation in hash.go's hex()
	expectedBytes := make([]byte, expectedBufferLen)
	copy(expectedBytes, expectedPrefix)
	copy(expectedBytes[len(expectedPrefix):], hexEncodedData)
	expectedStr := string(expectedBytes)

	if gotHex := h.Hex(); gotHex != expectedStr {
		t.Errorf("Hash.Hex() mismatch.\nGot : %q\nWant: %q", gotHex, expectedStr)
	}
}

func TestHash_BigInt(t *testing.T) {
	hashData := make([]byte, HashLen)
	for i := 0; i < HashLen; i++ {
		hashData[i] = byte(i + 5) // Fill with some non-zero distinct values
	}
	var h Hash
	h.SetBytes(hashData)

	wantBigInt := new(big.Int).SetBytes(h.Bytes()) // Use h.Bytes() to get the exact [HashLen]byte content
	gotBigInt := h.BigInt()

	if gotBigInt.Cmp(wantBigInt) != 0 {
		t.Errorf("Hash.BigInt() = %s, want %s", gotBigInt.String(), wantBigInt.String())
	}
}

func TestHash_Length(t *testing.T) {
	var h Hash
	if length := h.Length(); length != HashLen {
		t.Errorf("Hash.Length() = %d, want %d", length, HashLen)
	}
}

func TestBytesToHash(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Hash
	}{
		// Test cases are similar to SetBytes, as BytesToHash uses SetBytes internally
		{
			name:  "nil bytes",
			input: nil,
			want:  Hash{},
		},
		{
			name:  "shorter than HashLen",
			input: []byte{0x01, 0x02, 0x03},
			want: func() Hash {
				var wantHash Hash
				copy(wantHash[:], []byte{0x01, 0x02, 0x03})
				return wantHash
			}(),
		},
		{
			name:  "exact HashLen",
			input: bytes.Repeat([]byte{0xAA}, HashLen),
			want: func() Hash {
				var wantHash Hash
				copy(wantHash[:], bytes.Repeat([]byte{0xAA}, HashLen))
				return wantHash
			}(),
		},
		{
			name:  "longer than HashLen",
			input: append(bytes.Repeat([]byte{0xBB}, 5), bytes.Repeat([]byte{0xCC}, HashLen)...),
			want: func() Hash {
				var wantHash Hash
				copy(wantHash[:], bytes.Repeat([]byte{0xCC}, HashLen)) // BytesToHash -> SetBytes takes last HashLen
				return wantHash
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToHash(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesToHash() with input %x = %x, want %x", tt.input, got, tt.want)
			}
		})
	}
}

func TestBigIntToHash(t *testing.T) {
	valBytes := bytes.Repeat([]byte{0x05}, 8) // Shorter than HashLen
	val := new(big.Int).SetBytes(valBytes)
	hashFromBigInt := BigIntToHash(val)
	expectedHash := BytesToHash(val.Bytes()) // BytesToHash will handle left-padding with zeros
	if !reflect.DeepEqual(hashFromBigInt, expectedHash) {
		t.Errorf("BigIntToHash() for 8 bytes = %x, want %x", hashFromBigInt, expectedHash)
	}

	largeValBytes := bytes.Repeat([]byte{0x06}, HashLen+5) // Longer than HashLen
	largeVal := new(big.Int).SetBytes(largeValBytes)
	hashFromLargeBigInt := BigIntToHash(largeVal)
	// BigIntToHash -> BytesToHash -> SetBytes will take the last HashLen bytes of largeVal.Bytes()
	expectedBytesForHash := largeVal.Bytes()[len(largeVal.Bytes())-HashLen:]
	expectedHashFromLarge := BytesToHash(expectedBytesForHash)
	if !reflect.DeepEqual(hashFromLargeBigInt, expectedHashFromLarge) {
		t.Errorf("BigIntToHash() with large int (%d bytes) = %x, want %x", len(largeVal.Bytes()), hashFromLargeBigInt, expectedHashFromLarge)
	}
}

func TestHexToHash(t *testing.T) {
	tests := []struct {
		name   string
		hexStr string
		want   Hash
		panics bool
	}{
		{"valid hex with 0x", "0x" + strings.Repeat("1a", HashLen), BytesToHash(bytes.Repeat([]byte{0x1a}, HashLen)), false},
		{"valid hex no 0x", strings.Repeat("2b", HashLen), BytesToHash(bytes.Repeat([]byte{0x2b}, HashLen)), false},
		{"shorter hex", "0x112233", BytesToHash([]byte{0x11, 0x22, 0x33}), false},
		{"longer hex", "0x" + strings.Repeat("FF", HashLen+2), BytesToHash(bytes.Repeat([]byte{0xFF}, HashLen)), false}, // decode gives HashLen+2 bytes, BytesToHash takes last HashLen
		{"empty hex after prefix", "0x", BytesToHash([]byte{}), false},
		{"empty hex string", "", BytesToHash([]byte{}), false},
		{"invalid hex chars", "0xGGHHII", BytesToHash([]byte{0x00, 0x00, 0x00}), false}, // decode turns invalid to 0
		{"odd length hex with 0x", "0xabc", Hash{}, true},                               // decode panics
		{"odd length hex no 0x", "abc", Hash{}, true},                                  // decode panics
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("HexToHash(%q) did not panic as expected", tt.hexStr)
					}
				}()
			}
			got := HexToHash(tt.hexStr)
			if !tt.panics && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToHash(%q) = %x, want %x", tt.hexStr, got.Bytes(), tt.want.Bytes())
			}
		})
	}
}

func TestCXIDToHash(t *testing.T) {
	tests := []struct {
		name    string
		cxidStr string
		want    Hash
		panics  bool
	}{
		{"valid cxid with 1cx", "1cx" + strings.Repeat("3c", HashLen), BytesToHash(bytes.Repeat([]byte{0x3c}, HashLen)), false},
		{"valid hex no 1cx (decoded as hex)", strings.Repeat("4d", HashLen), BytesToHash(bytes.Repeat([]byte{0x4d}, HashLen)), false},
		{"odd length after 1cx strip", "1cx" + "abc", Hash{}, true}, // decode panics
		{"odd length no 1cx (decoded as hex)", "abc", Hash{}, true}, // decode panics
		{"empty after 1cx", "1cx", BytesToHash([]byte{}), false},
		{"longer cxid", "1cx" + strings.Repeat("EE", HashLen+3), BytesToHash(bytes.Repeat([]byte{0xEE}, HashLen)), false}, // decode gives HashLen+3 bytes, BytesToHash takes last HashLen
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("CXIDToHash(%q) did not panic as expected", tt.cxidStr)
					}
				}()
			}
			got := CXIDToHash(tt.cxidStr)
			if !tt.panics && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CXIDToHash(%q) = %x, want %x", tt.cxidStr, got.Bytes(), tt.want.Bytes())
			}
		})
	}
}

func TestStringToHash(t *testing.T) {
	tests := []struct {
		name   string
		strVal string
		want   Hash
	}{
		{"with 1cx prefix", "1cxHelloWorld", BytesToHash([]byte("HelloWorld"))}, // BytesToHash handles padding/truncation
		{"no 1cx prefix", "RawBytesData", BytesToHash([]byte("RawBytesData"))},
		{"shorter string", "short", BytesToHash([]byte("short"))},
		{
			"longer string with 1cx",
			"1cx" + strings.Repeat("A", HashLen+5),
			BytesToHash([]byte(strings.Repeat("A", HashLen))), // StringToHash -> SetBytes takes last HashLen
		},
		{
			"longer string no 1cx",
			strings.Repeat("B", HashLen+3),
			BytesToHash([]byte(strings.Repeat("B", HashLen))), // StringToHash -> SetBytes takes last HashLen
		},
		{"empty after 1cx", "1cx", BytesToHash([]byte{})},
		{"empty string", "", BytesToHash([]byte{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringToHash(tt.strVal)
			// Construct want by applying the same logic as StringToHash for clarity
			// (though tt.want already does this via BytesToHash)
			expectedRawBytes := []byte(tt.strVal)
			if strings.HasPrefix(tt.strVal, "1cx") {
				expectedRawBytes = []byte(tt.strVal[3:])
			}
			// BytesToHash internally calls SetBytes, which handles truncation/padding.
			wantHash := BytesToHash(expectedRawBytes)

			if !reflect.DeepEqual(got, wantHash) {
				t.Errorf("StringToHash(%q) = %x, want %x (derived from %x)", tt.strVal, got.Bytes(), wantHash.Bytes(), expectedRawBytes)
			}
			// Simpler check against pre-calculated want
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToHash(%q) = %x, want (pre-calc) %x", tt.strVal, got.Bytes(), tt.want.Bytes())
			}
		})
	}
}