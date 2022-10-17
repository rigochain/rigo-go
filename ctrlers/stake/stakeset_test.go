package stake_test

import (
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

var testGovRules = &gov.GovRules{
	Version:           0,
	MaxValidatorCnt:   21,
	RewardDelayBlocks: 10,
	AmountPerPower:    big.NewInt(1000),
	RewardPerPower:    big.NewInt(10),
}

func newStakeSet() *stake.StakeSet {
	w0 := client.NewWallet([]byte("1"))
	return stake.NewStakeSet(w0.Address(), w0.GetPubKey())
}

type stakesMap map[types.AcctKey][]*stake.Stake

func randomStakesTo(sset *stake.StakeSet, n int) stakesMap {
	var addr types.Address
	retStakeMap := make(stakesMap)

	stakes := make([]*stake.Stake, n)
	for i, _ := range stakes {
		if addr == nil || libs.RandInt63()%3 == 0 {
			addr = libs.RandAddress()
		}

		stakes[i] = stake.NewStakeWithAmount(
			addr,
			libs.RandBigIntN(new(big.Int).Quo(testGovRules.MaxStakeAmount(), big.NewInt(int64(n)))), // amount
			libs.RandInt63n(1_000_000), // height
			libs.RandHexBytes(32),      //txhash
			testGovRules)
		// fmt.Println(addr, stakes[i].Amount, stakes[i].Power)

		retStakeMap[types.ToAcctKey(addr)] = append(retStakeMap[types.ToAcctKey(addr)], stakes[i])
	}
	if err := sset.AppendStake(stakes...); err != nil {
		panic(err)
	}
	return retStakeMap
}

func TestStakeSet(t *testing.T) {
	stakesCnt := int(libs.RandInt63n(10000))
	stakeSet := newStakeSet()
	stakesMapByOwner := randomStakesTo(stakeSet, stakesCnt)
	require.NotNil(t, stakesMapByOwner)
	require.True(t, 0 < len(stakesMapByOwner))

	require.Equal(t, stakesCnt, stakeSet.StakesLen())
	require.Equal(t, stakeSet.SumPower(), stakeSet.TotalPower)
	require.Equal(t, stakeSet.SumAmount(), stakeSet.TotalAmount)

	sumPower := int64(0)
	sumAmt := big.NewInt(0)

	for acctKey, _ := range stakesMapByOwner {
		addr := acctKey.Address()
		stakes := stakeSet.StakesOf(addr)
		require.NotNil(t, stakes)

		powerOf := stakeSet.PowerOf(addr)
		_powerOf := int64(0)
		for _, s0 := range stakes {
			_powerOf += s0.Power
		}
		require.Equal(t, powerOf, _powerOf)

		sumPower += powerOf
		sumAmt = new(big.Int).Add(sumAmt, stakeSet.AmountOf(addr))
	}
	require.Equal(t, stakeSet.TotalPower, sumPower)
	require.Equal(t, stakeSet.TotalAmount, sumAmt)

	totalPower0 := stakeSet.TotalPower
	totalAmount0 := stakeSet.TotalAmount
	removedPower := int64(0)
	removedAmount := big.NewInt(0)
	removeCnt := rand.Intn(stakesCnt-1) + 1
	require.True(t, 0 < removeCnt)

	for i := 0; i < removeCnt; i++ {
		ridx := rand.Intn(stakesCnt - i)
		s0 := stakeSet.DelStakeByIdx(ridx)
		removedPower += s0.Power
		removedAmount = new(big.Int).Add(removedAmount, s0.Amount)
	}

	require.Equal(t, stakesCnt-removeCnt, stakeSet.StakesLen())
	require.Equal(t, totalPower0-removedPower, stakeSet.TotalPower)
	require.Equal(t, new(big.Int).Sub(totalAmount0, removedAmount), stakeSet.TotalAmount)
}

//func TestStakeOrder(t *testing.T) {
//	stakeSet := newStakeSet()
//	randomStakesTo(stakeSet, int(libs.RandInt63n(10000)))
//
//	var preStake *stake.Stake = nil
//	for i := 0; i < stakeSet.StakesLen(); i++ {
//		s := stakeSet.GetStake(i)
//		require.NotNil(t, s)
//
//		if preStake != nil {
//			require.Truef(t, preStake.StartHeight <= s.StartHeight,
//				fmt.Sprintf("pre stake height:%v, current stake height:%v", preStake.StartHeight, s.StartHeight))
//		}
//		preStake = s
//	}
//}

func TestApplyReward(t *testing.T) {
	validator := newStakeSet()
	randomStakesTo(validator, int(libs.RandInt63n(10000)))

	// first reward
	reward0 := validator.ApplyReward()
	require.Equal(t, reward0, validator.TotalReward)
	require.Equal(t, validator.SumReward(), validator.TotalReward)

	// second reward
	reward1 := validator.ApplyReward()
	require.Equal(t, reward0, reward1)
	require.Equal(t, new(big.Int).Add(reward0, reward1), validator.TotalReward)
	require.Equal(t, validator.SumReward(), validator.TotalReward)

	require.True(t, reward0.Sign() > 0)
	require.True(t, reward1.Sign() > 0)
	require.True(t, validator.TotalReward.Sign() > 0)
}

func BenchmarkApplyReward(b *testing.B) {
	val0 := newStakeSet()
	randomStakesTo(val0, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := val0.ApplyReward()
		require.True(b, rewarded.Sign() > 0)
	}
}
