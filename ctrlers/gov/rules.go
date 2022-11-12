package gov

import (
	"github.com/golang/protobuf/proto"
	"math/big"
)

type GovRule struct {
	Version            int64    `json:"version"`
	MaxValidatorCnt    int64    `json:"maxValidatorCnt"`
	AmountPerPower     *big.Int `json:"amountPerPower"`
	RewardPerPower     *big.Int `json:"rewardPerPower"`
	LazyRewardBlocks   int64    `json:"lazyRewardBlocks"`
	LazyApplyingBlocks int64    `json:"lazyApplyingBlocks"`
}

func DefaultGovRule() *GovRule {
	return &GovRule{
		Version:            0,
		MaxValidatorCnt:    21,
		AmountPerPower:     big.NewInt(1_000000000_000000000),
		RewardPerPower:     big.NewInt(1_000000000),
		LazyRewardBlocks:   10,
		LazyApplyingBlocks: 10,
	}
}

func DecodeGovRule(bz []byte) (*GovRule, error) {
	pm := &GovRuleProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return nil, err
	}
	return fromProto(pm), nil
}

func (gr *GovRule) Encode() ([]byte, error) {
	pm := toProto(gr)
	return proto.Marshal(pm)
}

func fromProto(pm *GovRuleProto) *GovRule {
	return &GovRule{
		Version:            pm.Version,
		MaxValidatorCnt:    pm.MaxValidatorCnt,
		AmountPerPower:     new(big.Int).SetBytes(pm.XAmountPerPower),
		RewardPerPower:     new(big.Int).SetBytes(pm.XRewardPerPower),
		LazyRewardBlocks:   pm.LazyRewardBlocks,
		LazyApplyingBlocks: pm.LazyApplyingBlocks,
	}
}

func toProto(gr *GovRule) *GovRuleProto {
	a := &GovRuleProto{
		Version:            gr.Version,
		MaxValidatorCnt:    gr.MaxValidatorCnt,
		XAmountPerPower:    gr.AmountPerPower.Bytes(),
		XRewardPerPower:    gr.RewardPerPower.Bytes(),
		LazyRewardBlocks:   gr.LazyRewardBlocks,
		LazyApplyingBlocks: gr.LazyApplyingBlocks,
	}
	return a
}
