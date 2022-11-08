package types

import (
	"math/big"
)

type IGovRules interface {
	GetMaxValidatorCount() int64
	GetLazyRewardBlocks() int64
	GetLazyApplyingBlocks() int64
	MaxStakeAmount() *big.Int
	MaxTotalPower() int64
	AmountToPower(*big.Int) int64
	PowerToAmount(int64) *big.Int
	PowerToReward(int64) *big.Int
}

const (
	PROPOSAL_GOVRULES = 1 + iota
	PROPOSAL_OTHERS
)

type IProposable interface {
	ID() int64
	Type() int32
	Encode() ([]byte, error)
	Decode([]byte) error

	GetStartVotingHeight() int64
	GetLastVotingHeight() int64
	GetApplyingHeight() int64
	GetMajorityPower() int64
	GetOption(int) interface{}
	GetVote(int) int64
}
