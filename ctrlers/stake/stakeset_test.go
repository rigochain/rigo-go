package stake_test

import (
	"fmt"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

var govRules = &gov.GovRules{
	Version:        0,
	AmountPerPower: big.NewInt(1000),
	RewardPerPower: big.NewInt(10),
}

var MAXSTAKEAMT = govRules.MaxStakeAmount()

func newStakeSet() *stake.StakeSet {
	w0 := client.NewWallet([]byte("1"))
	return stake.NewStakeSet(w0.Address(), w0.GetPubKey())
}

func newValidators(n int) stake.StakeSetArray {
	vals := make(stake.StakeSetArray, n)
	for i, _ := range vals {
		vals[i] = newStakeSet()
	}
	return vals
}

func newStakesTo(sset *stake.StakeSet, n int) {
	stakes := make([]*stake.Stake, n)
	for i, _ := range stakes {
		stakes[i] = stake.NewStakeWithAmount(
			sset.Owner,
			libs.RandBigIntN(MAXSTAKEAMT),
			libs.RandInt63n(10_000_0000),
			libs.RandHexBytes(32),
			govRules)
	}
	if err := sset.AppendStake(stakes...); err != nil {
		panic(err)
	}
}

func TestNewStake(t *testing.T) {
	w0 := client.NewWallet([]byte("1"))
	addr0 := w0.Address()
	stakeSet := stake.NewStakeSet(addr0, w0.GetPubKey())

	s0 := stake.NewStakeWithAmount(
		addr0,
		big.NewInt(1000), 1, nil,
		govRules)
	require.True(t, s0.Power > int64(0))
	require.Equal(t, govRules.AmountToPower(big.NewInt(1000)), s0.Power)
	require.True(t, s0.BlockReward.Sign() > 0)
	require.Equal(t, govRules.PowerToReward(s0.Power), s0.BlockReward)
	require.NoError(t, stakeSet.AppendStake(s0))

	s1 := stake.NewStakeWithAmount(
		addr0,
		big.NewInt(1000), 3, nil,
		govRules)
	require.NoError(t, stakeSet.AppendStake(s1))

	require.Equal(t, big.NewInt(2000), stakeSet.TotalAmount)
	require.Equal(t, 2, stakeSet.StakesLen())

	require.NotNil(t, stakeSet.PopStake())
	require.Equal(t, 1, stakeSet.StakesLen())
	require.Equal(t, big.NewInt(1000), stakeSet.TotalAmount)

	require.Nil(t, stakeSet.DelStake(1))
	require.Equal(t, 1, stakeSet.StakesLen())
	require.Equal(t, big.NewInt(1000), stakeSet.TotalAmount)
}

func TestStakeOrder(t *testing.T) {
	stakeSet := newStakeSet()
	newStakesTo(stakeSet, int(libs.RandInt63n(10000)))

	var preStake *stake.Stake = nil
	for i := 0; i < stakeSet.StakesLen(); i++ {
		s := stakeSet.GetStake(i)
		require.NotNil(t, s)

		if preStake != nil {
			require.Truef(t, preStake.StartHeight <= s.StartHeight,
				fmt.Sprintf("pre stake height:%v, current stake height:%v", preStake.StartHeight, s.StartHeight))
		}
		preStake = s
	}
}

func TestReward(t *testing.T) {
	validator := newStakeSet()
	newStakesTo(validator, int(libs.RandInt63n(10000)))

	reward0 := validator.ApplyReward()
	sumReward0 := big.NewInt(0)
	for i := 0; i < validator.StakesLen(); i++ {
		s := validator.GetStake(i)
		require.NotNil(t, s)

		sumReward0 = new(big.Int).Add(sumReward0, s.Reward)
	}

	require.Equal(t, reward0, validator.TotalReward)
	require.Equal(t, sumReward0, validator.TotalReward)

	reward1 := validator.ApplyReward()
	sumReward1 := big.NewInt(0)
	for i := 0; i < validator.StakesLen(); i++ {
		s := validator.GetStake(i)
		require.NotNil(t, s)

		sumReward1 = new(big.Int).Add(sumReward1, s.Reward)
	}
	require.Equal(t, reward0, reward1)
	require.Equal(t, reward1.Mul(reward0, big.NewInt(2)), validator.TotalReward)
	require.Equal(t, sumReward1, validator.TotalReward)

	require.True(t, reward0.Sign() > 0)
	require.True(t, reward1.Sign() > 0)
	require.True(t, sumReward0.Sign() > 0)
	require.True(t, sumReward1.Sign() > 0)
	require.True(t, validator.TotalReward.Sign() > 0)

}

func TestStakeSetList(t *testing.T) {
	vals := newValidators(21)
	for _, vset := range vals {
		newStakesTo(vset, 10000)
	}

	for i := 0; i < vals.Len(); i++ {
		vals[i].ApplyReward()
	}

	for i := 0; i < vals.Len(); i++ {
		require.Equal(t, vals[i].TotalAmount, vals[i].SumAmount())
		require.Equal(t, vals[i].TotalPower, vals[i].SumPower())
		require.Equal(t, vals[i].TotalReward, vals[i].SumReward())
	}
}

func BenchmarkApplyReward(b *testing.B) {
	val0 := newStakeSet()
	newStakesTo(val0, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := val0.ApplyReward()
		require.True(b, rewarded.Sign() > 0)
	}
}
