package types

import (
	"github.com/holiman/uint256"
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

// Simplest Asset Unit (SAU)

var (
	//XCOsau, _ = new(big.Int).SetString(XCO, 10)
	XCOsau = uint256.MustFromDecimal(XCO)
)

// Coint to SAU
func ToSAU(n uint64) *uint256.Int {
	return new(uint256.Int).Mul(uint256.NewInt(n), XCOsau)
}

// from sau to COIN and Remain
func FromSAU(sau *uint256.Int) (uint64, uint64) {
	r := new(uint256.Int)
	q, r := new(uint256.Int).DivMod(sau, XCOsau, r)
	return q.Uint64(), r.Uint64()
}
