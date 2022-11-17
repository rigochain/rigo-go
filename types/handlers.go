package types

import (
	"encoding/json"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/types"
	"math/big"
)

type ILedgerHandler interface {
	Query(*QueryData) (json.RawMessage, xerrors.XError)
	Commit() ([]byte, int64, error)
	Close() error
}

type IStakeHandler interface {
	GetTotalAmount() *big.Int
	GetTotalPower() int64
	GetTotalPowerOf(types.Address) int64
	IsValidator(types.Address) bool
}

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
