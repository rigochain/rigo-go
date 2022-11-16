package gov

import (
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

var txctxProposal *trxs.TrxContext

func init() {
	bz, err := govRuleProposal.Encode()
	if err != nil {
		panic(err)
	}

	txctxProposal = makeProposalTrxCtx(stakerHandlerHelper.PickAddress(rand.Int()), libs.ZeroAddress(), big.NewInt(10), bz)
}

func TestVoting(t *testing.T) {
	if err := govCtrler.Validate(txctxProposal); err != nil {
		panic(err)
	} else if err := govCtrler.Execute(txctxProposal); err != nil {
		panic(err)
	}

	expectedTotalPower := int64(0)
	propo := govCtrler.FindProposals(txctxProposal.TxHash)
	require.NotNil(t, propo)
	require.Equal(t, expectedTotalPower, propo.GetVotesFor(0))

	voter0 := stakerHandlerHelper.PickAddress(rand.Int())
	txctx0 := makeVotingTrxCtx(voter0, txctxProposal.TxHash, 0)
	require.NotNil(t, txctx0)

	if err := govCtrler.Validate(txctx0); err != nil {
		panic(err)
	} else if err := govCtrler.Execute(txctx0); err != nil {
		panic(err)
	}

	expectedVoterPower := txctx0.StakeHandler.GetTotalPowerOf(voter0)
	expectedTotalPower += expectedVoterPower
	require.True(t, 0 < expectedTotalPower)
	require.Equal(t, expectedVoterPower, propo.GetVotesOf(voter0, 0))
	require.Equal(t, expectedTotalPower, propo.GetVotesFor(0))

	t.Log(expectedVoterPower, expectedTotalPower)

	voter1 := stakerHandlerHelper.PickAddress(rand.Int())
	txctx1 := makeVotingTrxCtx(voter1, txctxProposal.TxHash, 0)
	require.NotNil(t, txctx1)

	if err := govCtrler.Validate(txctx1); err != nil {
		panic(err)
	} else if err := govCtrler.Execute(txctx1); err != nil {
		panic(err)
	}

	expectedVoterPower = txctx1.StakeHandler.GetTotalPowerOf(voter1)
	expectedTotalPower += expectedVoterPower
	require.True(t, 0 < expectedTotalPower)
	require.Equal(t, expectedVoterPower, propo.GetVotesOf(voter1, 0))
	require.Equal(t, expectedTotalPower, propo.GetVotesFor(0))

	t.Log(expectedVoterPower, expectedTotalPower)

}
