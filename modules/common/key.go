package common

import "math/big"

const KeyLen = 16

var (
	KeyByte = 0xf8
)

type Key [KeyLen]byte

func (k *Key) SetBytes(b []byte) {
	if len(b) > len(k) {
		b = b[len(b)-KeyLen:]
	}
	copy(k[:], b)
}

func (k *Key) Bytes() []byte {
	return k[:]
}

func (k *Key) String() string {
	return string(k.cxid())
}

func (k Key) Hex() string {
	return string(k.hex())
}

func (k Key) CXID() string {
	return string(k.cxid())
}

func BytesToKey(b []byte) Key {
	var k Key
	k.SetBytes(b)
	return k
}

func BigIntToKey(b *big.Int) Key {
	return BytesToKey(b.Bytes())
}

func (k Key) BigInt() *big.Int {
	return new(big.Int).SetBytes(k.Bytes())
}

func HexToKey(s string) Key {
	if has0xPrefix(s) {
		s = s[2:]
	}

	b := decode(s)
	return BytesToKey(b)
}

func CXIDToKey(s string) Key {
	if has1cxPrefix(s) {
		s = s[3:]
	}

	b := decode(s)
	return BytesToKey(b)
}

func StringToKey(s string) Key {
	return BytesToKey([]byte(s))
}

func (k Key) hex() []byte {
	buf := make([]byte, len(k)*2+2)
	copy(buf[:2], []byte("0x"))
	encode(buf[2:], k[:])
	return buf
}

func (k Key) cxid() []byte {
	buf := make([]byte, len(k)*3+2)
	copy(buf[:3], []byte("1cx"))
	copy(buf[3:5], []byte{byte(KeyByte)})
	encode(buf[5:], k[:])
	return buf
}
