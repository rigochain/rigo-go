package types

import "math/big"

type ILedgerCtrler interface {
	Commit() ([]byte, int64, error)
	Close() error
}

type IStakeCtrler interface {
	GetTotalAmount() *big.Int
	GetTotalPower() int64
}
