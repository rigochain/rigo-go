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

type IGovHandler interface {
	Version() int64
	MaxValidatorCnt() int64
	MinValidatorStake() *uint256.Int
	RewardPerPower() *uint256.Int
	LazyRewardBlocks() int64 // todo: rename LazyrewardBlocks to UnbondingPeriodBlockcs(?)
	LazyApplyingBlocks() int64
	GasPrice() *uint256.Int
	MinTrxGas() uint64
	MaxTrxGas() uint64
	MaxBlockGas() uint64
	MinVotingPeriodBlocks() int64
	MaxVotingPeriodBlocks() int64
	MinSelfStakeRatio() int64
	MaxUpdatableStakeRatio() int64
	SlashRatio() int64
}

type IAccountHandler interface {
	FindOrNewAccount(types.Address, bool) *Account
	FindAccount(types.Address, bool) *Account
	Transfer(types.Address, types.Address, *uint256.Int, bool) xerrors.XError
	Reward(types.Address, *uint256.Int, bool) xerrors.XError
	ImmutableAcctCtrlerAt(int64) (IAccountHandler, xerrors.XError)
	SetAccountCommittable(*Account, bool) xerrors.XError
}

type IStakeHandler interface {
	Validators() ([]*abcitypes.Validator, int64)
	IsValidator(types.Address) bool
	TotalPowerOf(types.Address) int64
	SelfPowerOf(types.Address) int64
	DelegatedPowerOf(types.Address) int64
}

type IDelegatee interface {
	GetAddress() types.Address
	GetTotalPower() int64
	GetSelfPower() int64
}
