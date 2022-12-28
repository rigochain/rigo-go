package types

import (
	"math/big"
)

const (
	ASSET_DECIMAL int16  = 18
	SYM_SAU       string = "sau"
	SYM_CURRENCY  string = "XCO"
	SAU                  = "1"
	KSAU                 = "1000"
	MSAU                 = "1000000"
	GSAU                 = "1000000000"
	UXCO                 = "1000000000000"
	MiXCO                = "1000000000000000"
	XCO                  = "1000000000000000000"             // 10^18 sau
	KXCO                 = "1000000000000000000000"          // 10^21 sau
	MXCO                 = "1000000000000000000000000"       // 10^24 sau
	GXCO                 = "1000000000000000000000000000"    // 10^27 sau
	TXCO                 = "1000000000000000000000000000000" // 10^30 sau
)

var (
	XCOsau, _ = new(big.Int).SetString(XCO, 10)
)

func ToSAU(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), XCOsau)
}

func FromSAU(sau *big.Int) (int64, int64) {
	r := new(big.Int)
	q, r := new(big.Int).QuoRem(sau, XCOsau, r)
	return q.Int64(), r.Int64()
}
