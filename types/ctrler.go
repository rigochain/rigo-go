package types

import (
	"github.com/tendermint/tendermint/types"
	"math/big"
)

type ILedgerCtrler interface {
	Commit() ([]byte, int64, error)
	Close() error
}

type IStakeCtrler interface {
	GetTotalAmount() *big.Int
	GetTotalPower() int64
	GetTotalPowerOf(types.Address) int64
}
