package gov

import (
	"github.com/golang/protobuf/proto"
	"math/big"
)

type GovRules struct {
	Version        int64
	AmountPerPower *big.Int
	RewardPerPower *big.Int
}

func NewGovRules() *GovRules {
	return &GovRules{
		AmountPerPower: big.NewInt(0),
		RewardPerPower: big.NewInt(0),
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
