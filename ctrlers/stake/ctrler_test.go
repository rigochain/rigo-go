package stake_test

import (
	"errors"
	"fmt"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"path/filepath"
	"testing"
)

var (
	addr0 = libs.RandAddress()
	dbDir = filepath.Join(os.TempDir(), "stake-ctrler-test")
)

func createStakers(cnt int, ctrler *stake.StakeCtrler) error {
	vals := newStakers(cnt)
	for _, val := range vals {
		if ctrler.AddStaker(val) == nil {
			return errors.New("function AddStaker() returns nil")
		}
	}
	return nil
}

func newStakers(n int) stake.StakeSetArray {
	vals := make(stake.StakeSetArray, n)
	for i, _ := range vals {
		vals[i] = newStakeSet()
	}
	return vals
}

var (
	stakingMapByStaker = make(map[types.AcctKey]stakesMap)
)

func stakingRandom(ctrler *stake.StakeCtrler) error {
	for i := 0; i < ctrler.StakersLen(); i++ {
		staker := ctrler.GetStaker(i)

		stakingMapByStaker[types.ToAcctKey(staker.Owner)] = randomStakesTo(staker, 10000)
	}
	return nil
}

func TestCreateStakers(t *testing.T) {
	vals := newStakers(21)
	for _, val := range vals {
		randomStakesTo(val, 10000)
	}

	for i := 0; i < vals.Len(); i++ {
		require.Equal(t, vals[i].TotalAmount, vals[i].SumAmount())
		require.Equal(t, vals[i].TotalPower, vals[i].SumPower())
		require.Equal(t, testGovRules.AmountToPower(vals[i].TotalAmount), vals[i].TotalPower)
	}
}

func TestStakerOrder(t *testing.T) {
	os.RemoveAll(dbDir)

	ctrler, err := stake.NewStakeCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)
	require.NoError(t, createStakers(21, ctrler))
	require.NoError(t, stakingRandom(ctrler))
	require.Equal(t, 21, ctrler.StakersLen())

	_ = ctrler.UpdateValidators(testGovRules) // sort

	var preVal *stake.StakeSet
	for i := 0; i < ctrler.StakersLen(); i++ {
		staker := ctrler.GetStaker(i)
		if preVal != nil {
			require.Truef(t, preVal.TotalPower >= staker.TotalPower,
				fmt.Sprintf("invalid power(stake) order: pre: %v, curr: %v", preVal.TotalPower, staker.TotalPower))
		}
		preVal = staker
	}

	require.NoError(t, ctrler.Close())
}

func TestLazyUnstaking(t *testing.T) {
	os.RemoveAll(dbDir)

	ctrler, err := stake.NewStakeCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)
	require.NoError(t, createStakers(21, ctrler))
	require.NoError(t, stakingRandom(ctrler))
	require.Equal(t, 21, ctrler.StakersLen())

	_ = ctrler.UpdateValidators(testGovRules) // sort

	//ridx := libs.RandInt63n(int64(ctrler.StakersLen()))
	//ctrler.Apply()
	//for i := 0; i < ctrler.StakersLen(); i++ {
	//	staker := ctrler.GetStaker(i)
	//
	//}
}

func TestStakerReward(t *testing.T) {
	// todo: check if reward is wrong or not
}
