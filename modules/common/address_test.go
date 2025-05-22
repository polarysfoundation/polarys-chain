package common

import (
	"bytes"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestAddress_SetBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Address
	}{
		{
			name:  "nil bytes",
			input: nil,
			want:  Address{},
		},
		{
			name:  "empty bytes",
			input: []byte{},
			want:  Address{},
		},
		{
			name:  "shorter than AddrLen",
			input: []byte{0x01, 0x02, 0x03},
			want: func() Address {
				var wantAddr Address
				inputBytes := []byte{0x01, 0x02, 0x03}
				copy(wantAddr[AddrLen-len(inputBytes):], inputBytes)
				return wantAddr
			}(),
		},
		{
			name:  "exact AddrLen",
			input: bytes.Repeat([]byte{0xAA}, AddrLen),
			want: func() Address {
				var addr Address
				copy(addr[:], bytes.Repeat([]byte{0xAA}, AddrLen))
				return addr
			}(),
		},
		{
			name:  "longer than AddrLen",
			input: append(bytes.Repeat([]byte{0xBB}, 10), bytes.Repeat([]byte{0xCC}, AddrLen)...),
			want: func() Address {
				var addr Address
				copy(addr[:], bytes.Repeat([]byte{0xCC}, AddrLen))
				return addr
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Address
			got.SetBytes(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Address.SetBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_StringAndCXID(t *testing.T) {
	addrData := bytes.Repeat([]byte{0xAB}, AddrLen)
	var addr Address
	addr.SetBytes(addrData)

	expectedPrefix := []byte{'1', 'c', 'x', 0x00}
	hexData := make([]byte, AddrLen*2)
	encode(hexData, addrData) // Uses package's encode

	bufferLen := AddrLen*3 + 2 // As per current implementation in address.go
	expectedBytes := make([]byte, bufferLen)
	copy(expectedBytes, expectedPrefix)
	copy(expectedBytes[len(expectedPrefix):], hexData)
	// The rest of expectedBytes (trailing part) is already 0x00 due to make.

	expectedStr := string(expectedBytes)

	if gotStr := addr.String(); gotStr != expectedStr {
		t.Errorf("Address.String() mismatch.\nGot : %q (len %d)\nWant: %q (len %d)", gotStr, len(gotStr), expectedStr, len(expectedStr))
		if !bytes.Equal([]byte(gotStr), []byte(expectedStr)) {
			t.Errorf("Address.String() byte content mismatch.\nGot : %x\nWant: %x", []byte(gotStr), []byte(expectedStr))
		}
	}

	if gotCXID := addr.CXID(); gotCXID != expectedStr {
		t.Errorf("Address.CXID() mismatch.\nGot : %q (len %d)\nWant: %q (len %d)", gotCXID, len(gotCXID), expectedStr, len(expectedStr))
		if !bytes.Equal([]byte(gotCXID), []byte(expectedStr)) {
			t.Errorf("Address.CXID() byte content mismatch.\nGot : %x\nWant: %x", []byte(gotCXID), []byte(expectedStr))
		}
	}
}

func TestAddress_Hex(t *testing.T) {
	addrData := bytes.Repeat([]byte{0xBC}, AddrLen)
	var addr Address
	addr.SetBytes(addrData)

	expectedPrefix := []byte{'0', 'x'}
	hexData := make([]byte, AddrLen*2)
	encode(hexData, addrData) // Uses package's encode

	expectedBufferLen := AddrLen*2 + 2 // As per current implementation in address.go
	expectedBytes := make([]byte, expectedBufferLen)
	copy(expectedBytes, expectedPrefix)
	copy(expectedBytes[len(expectedPrefix):], hexData)
	expectedStr := string(expectedBytes)

	if gotHex := addr.Hex(); gotHex != expectedStr {
		t.Errorf("Address.Hex() mismatch.\nGot : %q\nWant: %q", gotHex, expectedStr)
	}
}

func TestAddress_BigInt(t *testing.T) {
	addrData := make([]byte, AddrLen)
	for i := 0; i < AddrLen; i++ {
		addrData[i] = byte(i + 1) // Fill with 01, 02, ..., AddrLen
	}
	var addr Address
	addr.SetBytes(addrData)

	// The address bytes are right-padded if shorter, so BigInt should reflect the actual bytes in Address.
	wantBigInt := new(big.Int).SetBytes(addr.Bytes()) // Use addr.Bytes() to get the exact [AddrLen]byte content
	gotBigInt := addr.BigInt()

	if gotBigInt.Cmp(wantBigInt) != 0 {
		t.Errorf("Address.BigInt() = %s, want %s", gotBigInt.String(), wantBigInt.String())
	}
}

func TestBytesToAddress(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  Address
	}{
		{
			name:  "shorter than AddrLen",
			input: []byte{0x01, 0x02, 0x03},
			want: func() Address {
				var wantAddr Address
				inputBytes := []byte{0x01, 0x02, 0x03}
				copy(wantAddr[AddrLen-len(inputBytes):], inputBytes)
				return wantAddr
			}(),
		},
		{
			name:  "exact AddrLen",
			input: bytes.Repeat([]byte{0xAA}, AddrLen),
			want: func() Address {
				var addr Address
				copy(addr[:], bytes.Repeat([]byte{0xAA}, AddrLen))
				return addr
			}(),
		},
		{
			name:  "longer than AddrLen",
			input: append(bytes.Repeat([]byte{0xBB}, 5), bytes.Repeat([]byte{0xCC}, AddrLen)...),
			want: func() Address {
				var addr Address
				copy(addr[:], bytes.Repeat([]byte{0xCC}, AddrLen))
				return addr
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToAddress(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigIntToAddress(t *testing.T) {
	valBytes := bytes.Repeat([]byte{0x05}, 8)
	val := new(big.Int).SetBytes(valBytes)
	addrFromBigInt := BigIntToAddress(val)

	expectedAddr := BytesToAddress(val.Bytes())
	if !reflect.DeepEqual(addrFromBigInt, expectedAddr) {
		t.Errorf("BigIntToAddress() for 8 bytes = %v, want %v", addrFromBigInt, expectedAddr)
	}

	largeValBytes := bytes.Repeat([]byte{0x06}, AddrLen+5)
	largeVal := new(big.Int).SetBytes(largeValBytes)
	addrFromLargeBigInt := BigIntToAddress(largeVal)

	expectedAddrFromLarge := BytesToAddress(largeVal.Bytes())
	if !reflect.DeepEqual(addrFromLargeBigInt, expectedAddrFromLarge) {
		t.Errorf("BigIntToAddress() with large int = %v, want %v", addrFromLargeBigInt, expectedAddrFromLarge)
	}
}

func TestHashToAddress(t *testing.T) {
	hashData := bytes.Repeat([]byte{0xDD}, HashLen) // HashLen is 32
	var h Hash
	h.SetBytes(hashData)

	addrFromHash := HashToAddress(h)

	expectedAddrBytes := hashData[HashLen-AddrLen:] // Last AddrLen bytes of the hash
	expectedAddr := BytesToAddress(expectedAddrBytes)

	if !reflect.DeepEqual(addrFromHash, expectedAddr) {
		t.Errorf("HashToAddress() = %x, want %x", addrFromHash.Bytes(), expectedAddr.Bytes())
	}
}

func TestHexToAddress(t *testing.T) {
	tests := []struct {
		name   string
		hexStr string
		want   Address
		panics bool
	}{
		{"valid hex with 0x", "0x" + strings.Repeat("1a", AddrLen), BytesToAddress(bytes.Repeat([]byte{0x1a}, AddrLen)), false},
		{"valid hex no 0x", strings.Repeat("2b", AddrLen), BytesToAddress(bytes.Repeat([]byte{0x2b}, AddrLen)), false},
		{"shorter hex", "0x112233", BytesToAddress([]byte{0x11, 0x22, 0x33}), false},
		{"longer hex", "0x" + strings.Repeat("FF", AddrLen+5), BytesToAddress(bytes.Repeat([]byte{0xFF}, AddrLen)), false},
		{"empty hex after prefix", "0x", BytesToAddress([]byte{}), false},
		{"empty hex string", "", BytesToAddress([]byte{}), false},
		{"invalid hex chars", "0xGGHHII", BytesToAddress([]byte{0x00, 0x00, 0x00}), false}, // decode turns invalid to 0
		{"odd length hex", "0xabc", Address{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("HexToAddress(%q) did not panic as expected", tt.hexStr)
					}
				}()
			}
			got := HexToAddress(tt.hexStr)
			if !tt.panics && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToAddress(%q) = %x, want %x", tt.hexStr, got.Bytes(), tt.want.Bytes())
			}
		})
	}
}

func TestCXIDToAddress(t *testing.T) {
	tests := []struct {
		name    string
		cxidStr string
		want    Address
		panics  bool
	}{
		{"valid cxid with 1cx", "1cx" + strings.Repeat("3c", AddrLen), BytesToAddress(bytes.Repeat([]byte{0x3c}, AddrLen)), false},
		{"valid hex no 1cx (decoded as hex)", strings.Repeat("4d", AddrLen), BytesToAddress(bytes.Repeat([]byte{0x4d}, AddrLen)), false},
		{"odd length after 1cx strip", "1cx" + "abc", Address{}, true},
		{"odd length no 1cx (decoded as hex)", "abc", Address{}, true},
		{"empty after 1cx", "1cx", BytesToAddress([]byte{}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("CXIDToAddress(%q) did not panic as expected", tt.cxidStr)
					}
				}()
			}
			got := CXIDToAddress(tt.cxidStr)
			if !tt.panics && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CXIDToAddress(%q) = %x, want %x", tt.cxidStr, got.Bytes(), tt.want.Bytes())
			}
		})
	}
}

func TestStringToAddress(t *testing.T) {
	tests := []struct {
		name   string
		strVal string
		want   Address
	}{
		{"with 1cx prefix", "1cxHelloWorld", BytesToAddress([]byte("HelloWorld"))},
		{"no 1cx prefix", "RawBytesData", BytesToAddress([]byte("RawBytesData"))},
		{"shorter string", "short", BytesToAddress([]byte("short"))},
		{"longer string with 1cx", "1cx" + strings.Repeat("A", AddrLen+5), BytesToAddress([]byte(strings.Repeat("A", AddrLen+5)))},
		{"longer string no 1cx", strings.Repeat("B", AddrLen+3), BytesToAddress([]byte(strings.Repeat("B", AddrLen+3)))},
		{"empty after 1cx", "1cx", BytesToAddress([]byte{})},
		{"empty string", "", BytesToAddress([]byte{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringToAddress(tt.strVal)
			// BytesToAddress internally calls SetBytes, which handles truncation/padding.
			// So, construct want by also using BytesToAddress for consistency in how truncation/padding is applied.
			expectedRawBytes := []byte(tt.strVal)
			if strings.HasPrefix(tt.strVal, "1cx") {
				expectedRawBytes = []byte(tt.strVal[3:])
			}
			wantAddr := BytesToAddress(expectedRawBytes)

			if !reflect.DeepEqual(got, wantAddr) {
				t.Errorf("StringToAddress(%q) = %x, want %x", tt.strVal, got.Bytes(), wantAddr.Bytes())
			}
		})
	}
}
