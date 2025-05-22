package common

import (
	"math/big"
)

const AddrLen = 15

type Address [AddrLen]byte

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddrLen:]
	}
	copy(a[AddrLen-len(b):], b)
}

func (a Address) Bytes() []byte {
	return a[:]
}

func (a Address) String() string {
	return string(a.cxid())
}

func (a Address) Hex() string {
	return string(a.hex())
}

func (a Address) CXID() string {
	return string(a.cxid())
}

func (a Address) BigInt() *big.Int {
	return new(big.Int).SetBytes(a.Bytes())
}

func BytesToAddress(b []byte) Address {
	var addr Address
	addr.SetBytes(b)
	return addr
}

func BigIntToAddress(n *big.Int) Address {
	return BytesToAddress(n.Bytes())
}

func HashToAddress(h Hash) Address {
	return BytesToAddress(h.Bytes())
}

func HexToAddress(s string) Address {
	if has0xPrefix(s) {
		s = s[2:]
	}
	b := decode(s)
	return BytesToAddress(b)
}

func CXIDToAddress(s string) Address {
	if has1cxPrefix(s) {
		s = s[3:]
	}
	b := decode(s)

	return BytesToAddress(b[:])
}

func StringToAddress(s string) Address {
	return CXIDToAddress(s)
}

func (a Address) hex() []byte {
	buf := make([]byte, len(a)*2+2)
	copy(buf[:2], []byte("0x"))
	encode(buf[2:], a[:])
	return buf
}

func (a Address) cxid() []byte {
	buf := make([]byte, len(a)*2+3)
	copy(buf[:3], []byte("1cx"))
	encode(buf[3:], a[:])
	return buf
}
