package common

import "math/big"

const HashLen = 32

type Hash [HashLen]byte

func (h *Hash) SetBytes(buf []byte) {
	if len(buf) > HashLen {
		buf = buf[len(buf)-HashLen:]
	}
	copy(h[:], buf)
}

func (h Hash) IsValid() bool {
	for _, b := range h {
		if b != 0 {
			return true
		}
	}
	return false
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) String() string {
	return string(h.cxid())
}

func (h Hash) Hex() string {
	return string(h.hex())
}

func (h Hash) CXID() string {
	return string(h.cxid())
}

func (h Hash) BigInt() *big.Int {
	return new(big.Int).SetBytes(h.Bytes())
}

func (h Hash) Length() int {
	return len(h)
}

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func StringToHash(s string) Hash {
	if has1cxPrefix(s) {
		s = s[3:]
	}

	return BytesToHash([]byte(s))
}

func BigIntToHash(n *big.Int) Hash {
	return BytesToHash(n.Bytes())
}

func HexToHash(s string) Hash {
	if has0xPrefix(s) {
		s = s[2:]
	}

	b := decode(s)
	return BytesToHash(b)
}

func CXIDToHash(s string) Hash {
	if has1cxPrefix(s) {
		s = s[3:]
	}

	b := decode(s)
	return BytesToHash(b)
}

func (h Hash) hex() []byte {
	buf := make([]byte, len(h)*2+2)
	copy(buf[:2], []byte("0x"))
	encode(buf[2:], h[:])
	return buf
}

func (h Hash) cxid() []byte {
	buf := make([]byte, len(h)*2+3)
	copy(buf[:3], []byte("1cx"))
	encode(buf[3:], h[:])
	return buf
}
