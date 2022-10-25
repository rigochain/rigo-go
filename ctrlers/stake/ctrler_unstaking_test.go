package stake_test

import (
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
	lastHeight += testGovRules.GetRewardDelayBlocks()
	err := stakeCtrler.ProcessFrozenStakesAt(lastHeight, acctCtrlerHelper)
	require.NoError(t, err)

	stakeCtrler.Commit()
	frozenStakes = stakeCtrler.GetFrozenStakes()
	require.Equal(t, 0, len(frozenStakes))
}
