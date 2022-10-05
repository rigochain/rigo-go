package gov

import (
	"fmt"
	"github.com/kysee/arcanus/types"
	"math/big"
)

var (
	sauXCO, _    = new(big.Int).SetString(types.XCO, 10)
	sauVOTEPOWER = new(big.Int).Mul(sauXCO, big.NewInt(1))
	sauREWARD, _ = new(big.Int).SetString(types.GSAU, 10)
)

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
