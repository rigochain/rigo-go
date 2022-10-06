package gov

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
	"math"
	"math/big"
)

type GovRules struct {
	Version        uint64   `json:"version"`
	AmountPerPower *big.Int `json:"amountPerPower"`
	RewardPerPower *big.Int `json:"rewardPerPower"`
}

func DecodeGovRules(bz []byte) (*GovRules, error) {
	pm := &GovRulesProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return nil, err
	}
	return fromProto(pm), nil
}

func fromProto(pm *GovRulesProto) *GovRules {
	return &GovRules{
		Version:        pm.Version,
		AmountPerPower: new(big.Int).SetBytes(pm.XAmountPerPower),
		RewardPerPower: new(big.Int).SetBytes(pm.XRewardPerPower),
	}
}

func toProto(gr *GovRules) *GovRulesProto {
	return &GovRulesProto{
		Version:         gr.Version,
		XAmountPerPower: gr.AmountPerPower.Bytes(),
		XRewardPerPower: gr.RewardPerPower.Bytes(),
	}
}

func (gr *GovRules) Encode() ([]byte, error) {
	pm := toProto(gr)
	return proto.Marshal(pm)
}

//
// implement interface IGovRules
//
func (gr *GovRules) AmountToPower(amt *big.Int) int64 {
	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, gr.AmountPerPower)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (gr *GovRules) PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), gr.AmountPerPower)
}

func (gr *GovRules) PowerToReward(power int64) *big.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), gr.RewardPerPower)
}

// MaxStakeAmount means the max of amount which could be deposited.
// When the type of voting power is `int64`and VP:XCO = 1:1,
// the MAXSTAKEsau becomes `9223372036854775807 000000000000000000` (~= 922경 XCO)
func (gr *GovRules) MaxStakeAmount() *big.Int {
	return new(big.Int).Mul(big.NewInt(math.MaxInt64), gr.AmountPerPower)
}

var _ types.IGovRules = (*GovRules)(nil)
