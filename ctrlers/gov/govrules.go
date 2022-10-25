package gov

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/big"
)

type GovRules struct {
	Version           int64    `json:"version"`
	MaxValidatorCnt   int64    `json:"maxValidatorCnt"`
	RewardDelayBlocks int64    `json:"rewardDelayBlocks"`
	AmountPerPower    *big.Int `json:"amountPerPower"`
	RewardPerPower    *big.Int `json:"rewardPerPower"`
}

func DefaultGovRules() *GovRules {
	return &GovRules{
		Version:           0,
		MaxValidatorCnt:   21,
		RewardDelayBlocks: 10,
		AmountPerPower:    big.NewInt(1_000000000_000000000),
		RewardPerPower:    big.NewInt(1_000000000),
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
		Version:           pm.Version,
		MaxValidatorCnt:   pm.MaxValidatorCnt,
		RewardDelayBlocks: pm.RewardDelayBlocks,
		AmountPerPower:    new(big.Int).SetBytes(pm.XAmountPerPower),
		RewardPerPower:    new(big.Int).SetBytes(pm.XRewardPerPower),
	}
}

func toProto(gr *GovRules) *GovRulesProto {
	return &GovRulesProto{
		Version:           gr.Version,
		MaxValidatorCnt:   gr.MaxValidatorCnt,
		RewardDelayBlocks: gr.RewardDelayBlocks,
		XAmountPerPower:   gr.AmountPerPower.Bytes(),
		XRewardPerPower:   gr.RewardPerPower.Bytes(),
	}
}

func (gr *GovRules) Encode() ([]byte, error) {
	pm := toProto(gr)
	return proto.Marshal(pm)
}

//
// implement interface IGovRules
//

func (gr *GovRules) GetMaxValidatorCount() int64 {
	return gr.MaxValidatorCnt
}

// MaxStakeAmount means the max of amount which could be deposited.
// tmtypes.MaxTotalVotingPower = int64(math.MaxInt64) / 8
// When the type of voting power is `int64`and VP:XCO = 1:1,
// the MAXSTAKEsau becomes `int64(math.MaxInt64) / 8 * 10^18` (~= 922경 XCO)
func (gr *GovRules) MaxStakeAmount() *big.Int {
	return new(big.Int).Mul(big.NewInt(tmtypes.MaxTotalVotingPower), gr.AmountPerPower)
}

func (gr *GovRules) MaxTotalPower() int64 {
	return tmtypes.MaxTotalVotingPower
}

func (gr *GovRules) GetRewardDelayBlocks() int64 {
	return gr.RewardDelayBlocks
}

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

var _ types.IGovRules = (*GovRules)(nil)
