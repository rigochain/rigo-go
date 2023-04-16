package types

import (
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"math/big"
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
	AmountPerPower() *big.Int
	RewardPerPower() *big.Int
	LazyRewardBlocks() int64
	LazyApplyingBlocks() int64
	MinTrxFee() *big.Int
	MinVotingPeriodBlocks() int64
	MaxVotingPeriodBlocks() int64

	// utility methods
	MaxStakeAmount() *big.Int
	MaxTotalPower() int64
	AmountToPower(*big.Int) int64
	PowerToAmount(int64) *big.Int
	PowerToReward(int64) *big.Int
}

type IAccountHelper interface {
	FindOrNewAccount(types.Address, bool) *Account
	FindAccount(types.Address, bool) *Account
	Transfer(types.Address, types.Address, *big.Int, bool) xerrors.XError
	Reward(types.Address, *big.Int, bool) xerrors.XError
}

type IStakeHelper interface {
	Validators() ([]*abcitypes.Validator, int64)
	IsValidator(types.Address) bool
	PowerOf(types.Address) int64
}
