package stake_test

import (
	"bytes"
	"fmt"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestStakingToSelfByTx(t *testing.T) {
	sumAmt := big.NewInt(0)
	sumPower := int64(0)

	for _, txctx := range stakingToSelfTrxCtxs {
		err := stakeCtrler.Apply(txctx)
		require.NoError(t, err)

		sumAmt.Add(sumAmt, txctx.Tx.Amount)
		sumPower += txctx.GovRules.AmountToPower(txctx.Tx.Amount)
	}

	require.Equal(t, sumAmt.String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.GetTotalPower())
}

func TestStakingByTx(t *testing.T) {
	sumAmt := stakeCtrler.GetTotalAmount()
	sumPower := stakeCtrler.GetTotalPower()

	for _, txctx := range stakingTrxCtxs {
		err := stakeCtrler.Apply(txctx)
		require.NoError(t, err)

		sumAmt.Add(sumAmt, txctx.Tx.Amount)
		sumPower += txctx.GovRules.AmountToPower(txctx.Tx.Amount)
	}

	require.Equal(t, sumAmt.String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.GetTotalPower())
}

func TestUnstakingByTx(t *testing.T) {
	sumAmt0 := stakeCtrler.GetTotalAmount()
	sumPower0 := stakeCtrler.GetTotalPower()
	sumUnstakingAmt := big.NewInt(0)
	sumUnstakingPower := int64(0)

	for _, txctx := range unstakingTrxCtxs {
		stakingTxHash := txctx.Tx.Payload.(*trxs.TrxPayloadUnstaking).TxHash

		err := stakeCtrler.Apply(txctx)
		require.NoError(t, err)

		stakingTxCtx := findStakingTxCtx(stakingTxHash)

		sumUnstakingAmt.Add(sumUnstakingAmt, stakingTxCtx.Tx.Amount)
		sumUnstakingPower += txctx.GovRules.AmountToPower(stakingTxCtx.Tx.Amount)
	}

	require.Equal(t, new(big.Int).Sub(sumAmt0, sumUnstakingAmt).String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower0-sumUnstakingPower, stakeCtrler.GetTotalPower())

	stakeCtrler.Commit()

	// test freezing reward
	frozenStakes := stakeCtrler.GetFrozenStakes()
	require.Equal(t, len(unstakingTrxCtxs), len(frozenStakes))

	sumFrozenAmount := big.NewInt(0)
	sumFrozenPower := int64(0)
	for _, s := range frozenStakes {
		sumFrozenAmount.Add(sumFrozenAmount, s.Amount)
		sumFrozenPower += s.Power
	}
	require.Equal(t, sumFrozenAmount.String(), sumUnstakingAmt.String())
	require.Equal(t, sumFrozenPower, sumUnstakingPower)

	// test lazy rewarding
	lastHeight += testGovRules.GetLazyRewardBlocks()
	err := stakeCtrler.ProcessFrozenStakesAt(lastHeight, acctCtrlerHelper)
	require.NoError(t, err)

	stakeCtrler.Commit()
	frozenStakes = stakeCtrler.GetFrozenStakes()
	require.Equal(t, 0, len(frozenStakes))

}

func TestUpdateValidators(t *testing.T) {

	valUpdates0 := stakeCtrler.UpdateValidators(int(testGovRules.GetMaxValidatorCount())) // sort
	require.True(t, len(valUpdates0) > 0)
	require.True(t, stakeCtrler.GetLastValidatorCnt() > 0)

	valUpdates1 := stakeCtrler.UpdateValidators(int(testGovRules.GetMaxValidatorCount())) // sort
	require.True(t, len(valUpdates1) == 0)
	require.True(t, stakeCtrler.GetLastValidatorCnt() > 0)

	//
	// check validatorUpdates
	// todo: check that validators is correct or not
	//
	for i := 0; i < len(valUpdates0); i++ {
		found := false
		for j := 0; j < stakeCtrler.GetLastValidatorCnt(); j++ {
			val := stakeCtrler.GetDelegatee(j)
			require.NotNil(t, val)
			if valUpdates0[i].Power == val.GetTotalPower() &&
				bytes.Compare(valUpdates0[i].PubKey.GetSecp256K1(), val.PubKey) == 0 {
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

	var preDelegatee *stake.Delegatee
	for i := 0; i < stakeCtrler.DelegateeLen(); i++ {
		delegatee := stakeCtrler.GetDelegatee(i)

		////
		//// check power of staker
		//_stakesMapByOwner, ok := testStakingResultMapByStaker[types.ToAcctKey(delegatee.Addr)]
		//require.True(t, ok)
		//
		//for owner, stakingRet := range _stakesMapByOwner {
		//	require.Equal(t, types.Address(owner[:]), stakingRet.Owner)
		//	require.Equal(t, stakingRet.Power, staker.SumPowerOf(owner[:]))
		//	require.Equal(t, stakingRet.Amt, staker.SumAmountOf(owner[:]))
		//}

		//
		// order of stakes
		if preDelegatee != nil {
			if preDelegatee.TotalPower <= delegatee.TotalPower {
				fmt.Println("break")
			}
			require.Truef(t, preDelegatee.TotalPower > delegatee.TotalPower,
				fmt.Sprintf("invalid power(stake) order: pre: %v, curr: %v", preDelegatee.TotalPower, delegatee.TotalPower))
		}
		preDelegatee = delegatee
	}
}

func TestFrozenReward(t *testing.T) {

}
