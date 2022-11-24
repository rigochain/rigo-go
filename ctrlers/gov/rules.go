package gov

import (
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/json"
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
		LazyRewardBlocks:   20,
		LazyApplyingBlocks: 10,
	}
}

func DecodeGovRule(bz []byte) (*GovRule, error) {
	ret := &GovRule{}
	if err := ret.Decode(bz); err != nil {
		return nil, err
	}
	return ret, nil
}

func (gr *GovRule) Decode(bz []byte) error {
	pm := &GovRuleProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return err
	}
	gr.fromProto(pm)
	return nil
}

func (gr *GovRule) Encode() ([]byte, error) {
	return proto.Marshal(gr.toProto())
}

func (gr *GovRule) fromProto(pm *GovRuleProto) {
	gr.Version = pm.Version
	gr.MaxValidatorCnt = pm.MaxValidatorCnt
	gr.AmountPerPower = new(big.Int).SetBytes(pm.XAmountPerPower)
	gr.RewardPerPower = new(big.Int).SetBytes(pm.XRewardPerPower)
	gr.LazyRewardBlocks = pm.LazyRewardBlocks
	gr.LazyApplyingBlocks = pm.LazyApplyingBlocks
}

func (gr *GovRule) toProto() *GovRuleProto {
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

func (gr *GovRule) MarshalJSON() ([]byte, error) {
	tm := &struct {
		Version            int64  `json:"version"`
		MaxValidatorCnt    int64  `json:"maxValidatorCnt"`
		AmountPerPower     string `json:"amountPerPower"`
		RewardPerPower     string `json:"rewardPerPower"`
		LazyRewardBlocks   int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks int64  `json:"lazyApplyingBlocks"`
	}{
		Version:            gr.Version,
		MaxValidatorCnt:    gr.MaxValidatorCnt,
		AmountPerPower:     gr.AmountPerPower.String(),
		RewardPerPower:     gr.RewardPerPower.String(),
		LazyRewardBlocks:   gr.LazyRewardBlocks,
		LazyApplyingBlocks: gr.LazyApplyingBlocks,
	}
	return json.Marshal(tm)
}

func (gr *GovRule) UnmarshalJSON(bz []byte) error {
	tm := &struct {
		Version            int64  `json:"version"`
		MaxValidatorCnt    int64  `json:"maxValidatorCnt"`
		AmountPerPower     string `json:"amountPerPower"`
		RewardPerPower     string `json:"rewardPerPower"`
		LazyRewardBlocks   int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks int64  `json:"lazyApplyingBlocks"`
	}{}

	if err := json.Unmarshal(bz, tm); err != nil {
		return err
	}

	amtPower, ok := new(big.Int).SetString(tm.AmountPerPower, 10)
	if !ok {
		return xerrors.New("amountPerPower is wrong")
	}
	rwdPower, ok := new(big.Int).SetString(tm.RewardPerPower, 10)
	if !ok {
		return xerrors.New("rewardPerPower is wrong")
	}
	gr.Version = tm.Version
	gr.MaxValidatorCnt = tm.MaxValidatorCnt
	gr.AmountPerPower = amtPower
	gr.RewardPerPower = rwdPower
	gr.LazyRewardBlocks = tm.LazyRewardBlocks
	gr.LazyApplyingBlocks = tm.LazyApplyingBlocks
	return nil
}
