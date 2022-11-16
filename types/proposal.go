package types

import (
	"github.com/tendermint/tendermint/types"
)

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
