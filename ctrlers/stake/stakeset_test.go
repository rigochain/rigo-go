package stake_test

import (
	"fmt"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func newStakeSet() *stake.StakeSet {
	w0 := client.NewWallet([]byte("1"))
	return stake.NewStakeSet(w0.Address(), w0.GetPubKey())
}

func newValidatorSet(n int) stake.StakeSetArray {
	valSet := make(stake.StakeSetArray, 21)
	for i, _ := range valSet {
		valSet[i] = newStakeSet()
	}
	return valSet
}

func newStakesTo(sset *stake.StakeSet, n int) {
	maxAmount, ok := new(big.Int).SetString(types.MAXSTAKEsau, 10)
	if !ok {
		panic("invalid max amount string")
	}
	stakes := make([]*stake.Stake, n)
	for i, _ := range stakes {
		stakes[i] = stake.NewStake(sset.Owner, libs.RandBigIntN(maxAmount), libs.RandInt63n(10_000_0000), libs.RandHexBytes(32))
	}
	if err := sset.AppendStake(stakes...); err != nil {
		panic(err)
	}
}

func TestStakeSet(t *testing.T) {
	w0 := client.NewWallet([]byte("1"))
	addr0 := w0.Address()
	stakeSet := stake.NewStakeSet(addr0, w0.GetPubKey())

	s0 := stake.NewStake(addr0, big.NewInt(1000), 1, nil)
	require.NoError(t, stakeSet.AppendStake(s0))

	s1 := stake.NewStake(addr0, big.NewInt(1000), 3, nil)
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
	valSet := newValidatorSet(21)
	for _, vset := range valSet {
		newStakesTo(vset, 10000)
	}

	for i := 0; i < valSet.Len(); i++ {
		valSet[i].ApplyReward()
	}

	// todo: check reward
}

func BenchmarkApplyReward(b *testing.B) {
	validator := newStakeSet()
	newStakesTo(validator, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := validator.ApplyReward()
		require.True(b, rewarded.Sign() > 0)
	}
}
