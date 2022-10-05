package gov

import (
	"fmt"
	"github.com/kysee/arcanus/types"
	"github.com/tendermint/tendermint/libs/log"
	"math/big"
)

var (
	sauXCO, _    = new(big.Int).SetString(types.XCO, 10)
	sauVOTEPOWER = new(big.Int).Mul(sauXCO, big.NewInt(1))
	sauREWARD, _ = new(big.Int).SetString(types.GSAU, 10)
)

type GovCtrler struct {
	rules  *GovRules
	logger log.Logger
}

func NewGovCtrler(jsonBlobRules []byte, logger log.Logger) (*GovCtrler, error) {
	if rules, err := DecodeGovRules(jsonBlobRules); err != nil {
		return nil, err
	} else {
		return &GovCtrler{rules: rules, logger: logger}, nil
	}
}

func (ctrler *GovCtrler) AmountToPower(amt *big.Int) int64 {
	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, ctrler.rules.AmountPerPower)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (ctrler *GovCtrler) PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), ctrler.rules.AmountPerPower)
}

func (ctrler *GovCtrler) PowerToReward(vp int64) *big.Int {
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return new(big.Int).Mul(big.NewInt(vp), ctrler.rules.RewardPerPower)
}

var _ types.IGovRules = (*GovCtrler)(nil)
