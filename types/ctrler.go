package types

import (
	"github.com/tendermint/tendermint/types"
	"math/big"
)

type ILedgerHandler interface {
	Commit() ([]byte, int64, error)
	Close() error
}

type IStakeHandler interface {
	GetTotalAmount() *big.Int
	GetTotalPower() int64
	GetTotalPowerOf(types.Address) int64
	IsValidator(types.Address) bool
}
