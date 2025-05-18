package common

var (
	hextable = "0123456789abcdef"
)

func has0xPrefix(s string) bool {
	return len(s) >= 2 && s[0] == '0' && s[1] == 'x'
}

func has1cxPrefix(s string) bool {
	return len(s) >= 3 && s[0] == '1' && s[1] == 'c' && s[2] == 'x'
}

func ConvertInterfaceSliceToByteSlice(data []any) []byte {
	byteSlice := make([]byte, len(data))
	for i, v := range data {
		byteSlice[i] = byte(v.(float64))
	}
	return byteSlice
}

func Equal(slice1, slice2 []byte) bool {
	// If lengths are different, slices are not equal
	if len(slice1) != len(slice2) {
		return false
	}

	// Compare each byte
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false // Found a mismatch
		}
	}

	// All bytes are equal
	return true
}

func EncodeToHex(data []byte) string {
	return string(hexEncoder(data))
}

func EncodeToCXID(data []byte) string {
	return string(cxidEncoder(data))
}

func Decode(s string) []byte {
	return decode(s)
}

func hexEncoder(data []byte) []byte {
	// Create a buffer with enough space for "0x" and the hexadecimal representation of the data
	buf := make([]byte, len(data)*2+2)
	copy(buf[:2], []byte("0x"))
	// Encode the data into hexadecimal and write it to the buffer starting at index 2
	encode(buf[2:], data)

	return buf
}

func cxidEncoder(data []byte) []byte {
	// Create a buffer with enough space for "0x" and the hexadecimal representation of the data
	buf := make([]byte, len(data)*3+2)
	copy(buf[:3], []byte("1cx"))
	// Encode the data into hexadecimal and write it to the buffer starting at index 2
	encode(buf[3:], data)

	return buf
}

func encode(dst, src []byte) int {
	j := 0
	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f]
		j += 2
	}
	return len(src) * 2
}

func decode(s string) []byte {
	// Create a byte slice to hold the decoded bytes
	b := make([]byte, len(s)/2)

	// Decode the hex string into bytes
	for i := 0; i < len(s); i += 2 {
		b[i/2] = (hexToByte(s[i]) << 4) | hexToByte(s[i+1])
	}

	return b
}

func hexToByte(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	default:
		return 0
	}
}

func Uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	b[0] = byte(n >> 56)
	b[1] = byte(n >> 48)
	b[2] = byte(n >> 40)
	b[3] = byte(n >> 32)
	b[4] = byte(n >> 24)
	b[5] = byte(n >> 16)
	b[6] = byte(n >> 8)
	b[7] = byte(n)
	return b
}

func BytesToUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	n := uint64(b[0]) << 56
	n |= uint64(b[1]) << 48
	n |= uint64(b[2]) << 40
	n |= uint64(b[3]) << 32
	n |= uint64(b[4]) << 24
	n |= uint64(b[5]) << 16
	n |= uint64(b[6]) << 8
	n |= uint64(b[7])
	return n
}
