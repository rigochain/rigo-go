package stake_test

import (
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

var (
	addr0 = libs.RandAddress()
	addr1 = libs.RandAddress()
)

func TestNewStake(t *testing.T) {
	amt := libs.RandBigIntN(govRuleHandlerHelper.MaxStakeAmount())
	s0 := stake.NewStakeWithAmount(
		addr0,
		addr1,
		amt, 1, nil,
		govRuleHandlerHelper)

	require.True(t, s0.Power > int64(0))
	require.Equal(t, govRuleHandlerHelper.AmountToPower(amt), s0.Power)
	require.True(t, s0.BlockRewardUnit.Sign() > 0)
	require.Equal(t, govRuleHandlerHelper.PowerToReward(s0.Power), s0.BlockRewardUnit)
	require.Equal(t, big.NewInt(0), s0.ReceivedReward)
}

func TestApplyRewardByStake(t *testing.T) {
	stakes := make([]*stake.Stake, 1000)

	for i := 0; i < 1000; i++ {
		amt := libs.RandBigIntN(new(big.Int).Div(govRuleHandlerHelper.MaxStakeAmount(), big.NewInt(1000)))
		txhash := libs.RandBytes(32)
		stakes[i] = stake.NewStakeWithAmount(
			addr0,
			addr1,
			amt, 1, txhash,
			govRuleHandlerHelper)
	}

	for i := 0; i < 1000; i++ {
		_ = stakes[i].ApplyReward()
		require.True(t, stakes[i].ReceivedReward.Sign() > 0)
		require.Equal(t, stakes[i].BlockRewardUnit, stakes[i].ReceivedReward)
	}

	for i := 0; i < 1000; i++ {
		_ = stakes[i].ApplyReward()
		expected := new(big.Int).Mul(stakes[i].BlockRewardUnit, big.NewInt(2))
		require.Equal(t, expected, stakes[i].ReceivedReward)
	}
}
