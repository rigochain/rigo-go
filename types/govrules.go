package types

import "math/big"

type IGovRules interface {
	GetMaxValidatorCount() int64
	GetRewardDelayBlocks() int64
	MaxStakeAmount() *big.Int
	MaxTotalPower() int64
	AmountToPower(*big.Int) int64
	PowerToAmount(int64) *big.Int
	PowerToReward(int64) *big.Int
}
