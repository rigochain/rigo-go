package gov

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
	"math/big"
)

type GovRules struct {
	version         int64
	amountPerPowner *big.Int
	rewardPerPower  *big.Int
}

func NewGovRules() *GovRules {
	return &GovRules{
		amountPerPowner: big.NewInt(0),
		rewardPerPower:  big.NewInt(0),
	}
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
		version:         pm.Version,
		amountPerPowner: new(big.Int).SetBytes(pm.XAmountPerPower),
		rewardPerPower:  new(big.Int).SetBytes(pm.XRewardPerPower),
	}
}

func toProto(gr *GovRules) *GovRulesProto {
	return &GovRulesProto{
		Version:         gr.version,
		XAmountPerPower: gr.amountPerPowner.Bytes(),
		XRewardPerPower: gr.rewardPerPower.Bytes(),
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
	_vp := new(big.Int).Quo(amt, gr.amountPerPowner)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (gr *GovRules) PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), gr.amountPerPowner)
}

func (gr *GovRules) PowerToReward(power int64) *big.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), gr.rewardPerPower)
}

var _ types.IGovRules = (*GovRules)(nil)
