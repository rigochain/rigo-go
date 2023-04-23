package types

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type ILedgerHandler interface {
	InitLedger(interface{}) xerrors.XError
	Commit() ([]byte, int64, xerrors.XError)
	Query(abcitypes.RequestQuery) ([]byte, xerrors.XError)
	Close() xerrors.XError
}

type IGovHelper interface {
	Version() int64
	MaxValidatorCnt() int64
	AmountPerPower() *uint256.Int
	RewardPerPower() *uint256.Int
	LazyRewardBlocks() int64
	LazyApplyingBlocks() int64
	MinTrxFee() *uint256.Int
	MinVotingPeriodBlocks() int64
	MaxVotingPeriodBlocks() int64

	// utility methods
	MaxStakeAmount() *uint256.Int
	MaxTotalPower() int64
	AmountToPower(*uint256.Int) int64
	PowerToAmount(int64) *uint256.Int
	PowerToReward(int64) *uint256.Int
}

type IAccountHelper interface {
	FindOrNewAccount(types.Address, bool) *Account
	FindAccount(types.Address, bool) *Account
	Transfer(types.Address, types.Address, *uint256.Int, bool) xerrors.XError
	Reward(types.Address, *uint256.Int, bool) xerrors.XError
}

type IStakeHelper interface {
	Validators() ([]*abcitypes.Validator, int64)
	IsValidator(types.Address) bool
	PowerOf(types.Address) int64
}
