package types

import "math/big"

type IGovRules interface {
	AmountToPower(*big.Int) int64
	PowerToAmount(int64) *big.Int
	PowerToReward(int64) *big.Int
	MaxStakeAmount() *big.Int
}
