package utils

import (
	"crypto/rand"
	"math/big"
)

func SecureRandomInt(min, max int) int {
	if min > max {
		min, max = max, min
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		panic(err)
	}
	return min + int(nBig.Int64())
}
