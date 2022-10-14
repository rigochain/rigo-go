package types

import "math/big"

type IGovRules interface {
	MaxValidatorCount() int32
	GetRewardDelayBlocks() int64
	MaxStakeAmount() *big.Int
	AmountToPower(*big.Int) int64
	PowerToAmount(int64) *big.Int
	PowerToReward(int64) *big.Int
}
