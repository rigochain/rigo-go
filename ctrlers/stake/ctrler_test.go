package stake_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

var (
	dbDir                        = filepath.Join(os.TempDir(), "stake-ctrler-test")
	testStakingResultMapByStaker = make(map[types.AcctKey]StakeResultsMap)
	testStakingResultMapByOwner  = make(StakeResultsMap)
)

func createDelegatees(ctrler *stake.StakeCtrler, cnt int) error {
	vals := newDelegatees(cnt)
	for _, val := range vals {
		if ctrler.AddDelegatee(val) == nil {
			return errors.New("function AddDelegatee() returns nil")
		}
	}
	return nil
}

func newDelegatees(n int) stake.DelegateeArray {
	vals := make(stake.DelegateeArray, n)
	for i, _ := range vals {
		vals[i] = newDelegatee()
	}
	return vals
}

func stakingRandom(ctrler *stake.StakeCtrler, maxCnt int) error {
	allowedPowerForStaker := (testGovRules.MaxTotalPower() - ctrler.GetTotalPower()) / int64(ctrler.DelegateeLen())

	for i := 0; i < ctrler.DelegateeLen(); i++ {
		staker := ctrler.GetDelegatee(i)

		rn := rand.Intn(maxCnt)
		_stakingResultMapByOwner, err := randomStakesTo(staker, rn, allowedPowerForStaker)
		if err != nil {
			return err
		}
		for k, _sr := range _stakingResultMapByOwner {
			if sret, ok := testStakingResultMapByOwner[k]; ok {
				if bytes.Compare(sret.Owner, _sr.Owner) != 0 {
					return errors.New("wrong owner")
				}
				sret.Power += _sr.Power
				sret.Amt = sret.Amt.Add(sret.Amt, _sr.Amt)
			} else {
				testStakingResultMapByOwner[k] = &StakingResult{
					Owner: k.Address(),
					Power: _sr.Power,
					Amt:   _sr.Amt,
				}
			}
		}
		testStakingResultMapByStaker[types.ToAcctKey(staker.Addr)] = _stakingResultMapByOwner
	}
	return nil
}

func TestCreateStakers(t *testing.T) {
	vals := newDelegatees(int(testGovRules.GetMaxValidatorCount()))
	for _, val := range vals {
		_, err := randomStakesTo(val, 10000, testGovRules.MaxTotalPower()-vals.SumTotalPower())
		require.NoError(t, err)
	}

	for i := 0; i < len(vals); i++ {
		require.Equal(t, vals[i].TotalAmount, vals[i].SumAmount())
		require.Equal(t, vals[i].TotalPower, vals[i].SumPower())
		require.Equal(t, testGovRules.AmountToPower(vals[i].TotalAmount), vals[i].TotalPower)
	}
}

func TestStakerStakes(t *testing.T) {
	os.RemoveAll(dbDir)

	ctrler, err := stake.NewStakeCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)

	stakerCnt := rand.Intn(1000)
	stakeCntOfStaker := rand.Intn(10000)
	require.NoError(t, createDelegatees(ctrler, stakerCnt))
	require.NoError(t, stakingRandom(ctrler, stakeCntOfStaker))
	require.Equal(t, stakerCnt, ctrler.DelegateeLen())

	valUpdates0 := ctrler.UpdateValidators(int(testGovRules.GetMaxValidatorCount())) // sort
	//
	// check validatorUpdates
	for i := 0; i < len(valUpdates0); i++ {
		found := false
		for j := 0; j < ctrler.GetLastValidatorCnt(); j++ {
			staker := ctrler.GetDelegatee(j)
			require.NotNil(t, staker)
			if valUpdates0[i].Power == staker.TotalPower &&
				bytes.Compare(valUpdates0[i].PubKey.GetSecp256K1(), staker.PubKey) == 0 {
				found = true
			}
		}

		if valUpdates0[i].Power == 0 {
			// because valUpdates0[i] is excluded, it MUST not exist in validator list.
			require.False(t, found)
		} else {
			// because valUpdates0[i] is added or changed, it MUST exist in validator list.
			require.True(t, found)
		}
	}

	var preVal *stake.Delegatee
	for i := 0; i < ctrler.DelegateeLen(); i++ {
		staker := ctrler.GetDelegatee(i)

		//
		// check power of staker
		_stakesMapByOwner, ok := testStakingResultMapByStaker[types.ToAcctKey(staker.Addr)]
		require.True(t, ok)

		for owner, stakingRet := range _stakesMapByOwner {
			require.Equal(t, types.Address(owner[:]), stakingRet.Owner)
			require.Equal(t, stakingRet.Power, staker.PowerOf(owner[:]))
			require.Equal(t, stakingRet.Amt, staker.AmountOf(owner[:]))
		}

		//
		// order of stakes
		if preVal != nil {
			if preVal.TotalPower <= staker.TotalPower {
				fmt.Println("break")
			}
			require.Truef(t, preVal.TotalPower > staker.TotalPower,
				fmt.Sprintf("invalid power(stake) order: pre: %v, curr: %v", preVal.TotalPower, staker.TotalPower))
		}
		preVal = staker
	}

	// add staking
	require.NoError(t, stakingRandom(ctrler, 1000))
	valUpdates2 := ctrler.UpdateValidators(int(testGovRules.GetMaxValidatorCount()))

	for i := 0; i < len(valUpdates2); i++ {
		found := false
		for j := 0; j < ctrler.GetLastValidatorCnt(); j++ {
			staker := ctrler.GetDelegatee(j)
			require.NotNil(t, staker)
			if valUpdates2[i].Power == staker.TotalPower &&
				bytes.Compare(valUpdates2[i].PubKey.GetSecp256K1(), staker.PubKey) == 0 {
				found = true
			}
		}

		if valUpdates2[i].Power == 0 {
			// because valUpdates0[i] is excluded, it MUST not exist in validator list.
			require.False(t, found)
		} else {
			// because valUpdates0[i] is added or changed, it MUST exist in validator list.
			require.True(t, found)
		}
	}

	preVal = nil
	for i := 0; i < ctrler.DelegateeLen(); i++ {
		staker := ctrler.GetDelegatee(i)

		//
		// check power of staker
		_stakesMapByOwner, ok := testStakingResultMapByStaker[types.ToAcctKey(staker.Addr)]
		require.True(t, ok)

		for owner, stakingRet := range _stakesMapByOwner {
			require.Equal(t, types.Address(owner[:]), stakingRet.Owner)
			require.Equal(t, stakingRet.Power, staker.PowerOf(owner[:]))
			require.Equal(t, stakingRet.Amt, staker.AmountOf(owner[:]))
		}

		//
		// order of stakes
		if preVal != nil {
			require.Truef(t, preVal.TotalPower > staker.TotalPower,
				fmt.Sprintf("invalid power(stake) order: pre: %v, curr: %v", preVal.TotalPower, staker.TotalPower))
		}
		preVal = staker
	}

	// someone's power is equal to the sum of power deposited to stakers by someone
	for k, sret := range testStakingResultMapByOwner {
		sumPower := int64(0)
		for i := 0; i < ctrler.DelegateeLen(); i++ {
			staker := ctrler.GetDelegatee(i)
			require.NotNil(t, staker)

			sumPower += staker.PowerOf(k.Address())
		}
		require.Equal(t, sumPower, sret.Power)
	}

	sumPower := int64(0)
	for i := 0; i < ctrler.DelegateeLen(); i++ {
		staker := ctrler.GetDelegatee(i)
		require.NotNil(t, staker)

		sumPower += staker.GetTotalPower()
	}
	require.True(t, sumPower > 0, "sum of power", sumPower)

	require.NoError(t, ctrler.Close())
}

func TestLazyUnstaking(t *testing.T) {
	//txctx := &trxs.TrxContext{
	//	Tx: &trxs.Trx{
	//		Version: 0,
	//		Time:    0,
	//		Nonce:   0,
	//		From:    nil,
	//		To:      nil,
	//		Amount:  nil,
	//		Gas:     nil,
	//		Type:    0,
	//		Payload: nil,
	//		Sig:     nil,
	//	},
	//	TxHash:       nil,
	//	Height:       0,
	//	Proposer:     nil,
	//	Exec:         false,
	//	SenderPubKey: nil,
	//	Sender:       nil,
	//	Receiver:     nil,
	//	NeedAmt:      nil,
	//	GasUsed:      nil,
	//	GovRules:     nil,
	//	Error:        nil,
	//}

}

func TestStakerReward(t *testing.T) {
	// todo: check if reward is wrong or not
}
