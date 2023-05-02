package stake_test

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	addr0 = types.RandAddress()
	addr1 = types.RandAddress()
)

func TestNewStake(t *testing.T) {
	amt := bytes.RandU256IntN(govHelper.MaxStakeAmount())
	s0 := stake.NewStakeWithAmount(
		addr0,
		addr1,
		amt, 1, nil,
		govHelper)

	require.True(t, s0.Power > int64(0))
	require.Equal(t, govHelper.AmountToPower(amt), s0.Power)
	require.True(t, s0.BlockRewardUnit.Sign() > 0)
	require.Equal(t, govHelper.PowerToReward(s0.Power), s0.BlockRewardUnit)
	require.Equal(t, uint256.NewInt(0), s0.ReceivedReward)
}

func TestApplyRewardByStake(t *testing.T) {
	stakes := make([]*stake.Stake, 1000)

	for i := 0; i < 1000; i++ {
		amt := bytes.RandU256IntN(new(uint256.Int).Div(govHelper.MaxStakeAmount(), uint256.NewInt(1000)))
		txhash := bytes.RandBytes(32)
		stakes[i] = stake.NewStakeWithAmount(
			addr0,
			addr1,
			amt, 1, txhash,
			govHelper)
	}

	for i := 0; i < 1000; i++ {
		_ = stakes[i].ApplyReward()
		require.True(t, stakes[i].ReceivedReward.Sign() > 0)
		require.Equal(t, stakes[i].BlockRewardUnit, stakes[i].ReceivedReward)
	}

	for i := 0; i < 1000; i++ {
		_ = stakes[i].ApplyReward()
		expected := new(uint256.Int).Mul(stakes[i].BlockRewardUnit, uint256.NewInt(2))
		require.Equal(t, expected, stakes[i].ReceivedReward)
	}
}
