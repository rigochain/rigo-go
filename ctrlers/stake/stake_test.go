package stake_test

import (
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewStake(t *testing.T) {
	amt := libs.RandBigIntN(MAXSTAKEAMT)
	s0 := stake.NewStakeWithAmount(
		addr0,
		amt, 1, nil,
		govRules)

	require.True(t, s0.Power > int64(0))
	require.Equal(t, govRules.AmountToPower(amt), s0.Power)
	require.True(t, s0.BlockReward.Sign() > 0)
	require.Equal(t, govRules.PowerToReward(s0.Power), s0.BlockReward)
}

func TestApplyReward(t *testing.T) {
	stakes := make([]*stake.Stake, 1000)
	amt := libs.RandBigIntN(MAXSTAKEAMT)
	s0 := stake.NewStakeWithAmount(
		addr0,
		amt, 1, nil,
		govRules)
}
