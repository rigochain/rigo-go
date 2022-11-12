package types

import (
	"github.com/tendermint/tendermint/types"
	"math/big"
)

type IGovRuleHandler interface {
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
	PROPOSAL_GOVRULE = 1 + iota
	PROPOSAL_OTHERS
)

type IProposable interface {
	ID() []byte
	Type() int32
	Encode() ([]byte, error)
	Decode([]byte) error

	GetStartVotingHeight() int64
	GetLastVotingHeight() int64
	GetApplyingHeight() int64
	GetMajorityPower() int64
	GetOption(int) interface{}
	GetVotesFor(int32) int64
	GetVotesOf(types.Address, int32) int64
	DoVote(types.Address, int64)
}
