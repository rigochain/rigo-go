package libs

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"math/big"
	mrand "math/rand"
)

func RandBytes(n int) []byte {
	bz := make([]byte, n)
	_, _ = rand.Read(bz)
	return bz
}

func ZeroBytes(n int) []byte {
	return make([]byte, n)
}

func RandHexBytes(n int) types.HexBytes {
	bz := RandBytes(n)
	return types.HexBytes(bz)
}

func RandHexString(n int) string {
	bz := RandBytes(n)
	return "0x" + hex.EncodeToString(bz)
}

func RandAddress() account.Address {
	return RandBytes(account.AddrSize)
}

func ZeroAddress() account.Address {
	return ZeroBytes(account.AddrSize)
}

func RandAddrKey() account.AcctKey {
	var k account.AcctKey
	copy(k[:], RandBytes(account.AddrSize))
	return k
}

func RandBigIntN(cap *big.Int) *big.Int {
	r, _ := rand.Int(rand.Reader, cap)
	return r
}

func RandBigInt() *big.Int {
	return new(big.Int).SetBytes(RandBytes(32))
}

func RandInt63() int64 {
	return mrand.Int63()
}

func RandInt63n(n int64) int64 {
	return mrand.Int63n(n)
}

func ClearBytes(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0x00
	}
}
