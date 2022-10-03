package types

import (
	"fmt"
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
	uXCO                 = "1000000000000"
	mXCO                 = "1000000000000000"
	XCO                  = "1000000000000000000"             // 10^18 sau
	KXCO                 = "1000000000000000000000"          // 10^21 sau
	MXCO                 = "1000000000000000000000000"       // 10^24 sau
	GXCO                 = "1000000000000000000000000000"    // 10^27 sau
	TXCO                 = "1000000000000000000000000000000" // 10^30 sau

	/*
	 * MAXSTAKEsau means the max of amount which could be deposited.
	 * When the type of voting power is `int64`and VP:XCO = 1:1,
	 * the MAXSTAKEsau becomes `9223372036854775807 000000000000000000` (~= 922경 XCO)
	 */
	MAXSTAKEsau = "9223372036854775807000000000000000000" // ~= 922경 XCO
)

var (
	sauXCO, _    = new(big.Int).SetString(XCO, 10)
	sauVOTEPOWER = new(big.Int).Mul(sauXCO, big.NewInt(1))
	sauREWARD, _ = new(big.Int).SetString(GSAU, 10)
)

func ToSAU(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), sauXCO)
}

func FromSAU(sau *big.Int) (int64, int64) {
	r := new(big.Int)
	q, r := new(big.Int).QuoRem(sau, sauXCO, r)
	return q.Int64(), r.Int64()
}

func AmountToPower(amt *big.Int) int64 {
	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, sauVOTEPOWER)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), sauVOTEPOWER)
}

func PowerToReward(vp int64) *big.Int {
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return new(big.Int).Mul(big.NewInt(vp), sauREWARD)
}
