package crypto

import (
	"hash"
	"log"
	"math/big"

	pec256 "github.com/polarysfoundation/pec-256"
	pm256 "github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
)

var c = pec256.PEC256()

func Pm256(b []byte) []byte {
	buf := make([]byte, 32)
	h := pm256.New256()
	h.Write(b)
	h.Sum(buf[:0])

	return buf
}

func GenerateKey() (pec256.PrivKey, pec256.PubKey) {
	priv, pub, _, err := c.GenerateKeyPair()
	if err != nil {
		log.Printf("error generating keys: %v", err)
		panic("error creating new keypair")
	}

	return priv, pub
}

func CreateAddress(a common.Address, n uint64, h common.Hash) common.Address {
	nonce := new(big.Int)
	nonce.SetUint64(n)

	data := make([]byte, len(nonce.Bytes())+len(a.Bytes())+len(h.Bytes()))
	data[0] = 0xff
	copy(data[1:], nonce.Bytes())
	copy(data[1+len(nonce.Bytes()):], a.Bytes())
	copy(data[1+len(nonce.Bytes())+len(a.Bytes()):], h.Bytes())

	return common.BytesToAddress(Pm256(data)[len(data)-common.AddrLen:])
}

func Sign(data common.Hash, priv pec256.PrivKey) (*big.Int, *big.Int, error) {
	return c.Sign(data.Bytes(), priv.BigInt())
}

func GenerateSharedKey(priv pec256.PrivKey) pec256.SharedKey {
	return c.SharedKey(priv)
}

func Verify(data common.Hash, r, s *big.Int, pub pec256.PubKey) (bool, error) {
	return c.Verify(data[:], r, s, pub.BigInt())
}

func NewPM256() hash.Hash {
	return pm256.New256()
}

func GetPubKey(priv pec256.PrivKey) pec256.PubKey {
	pub, _ := c.GetPubKey(priv)

	if !c.IsValidPubKey(pub.BigInt()) {
		panic("invalid pubkey")
	}

	return pub
}

func PubKeyToAddress(pub pec256.PubKey) common.Address {
	b := pub.Bytes()

	return common.BytesToAddress(Pm256(b)[len(b)-15:])
}
