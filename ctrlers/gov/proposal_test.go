package gov

import (
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/big"
	"math/rand"
	"testing"
)

type Case struct {
	txctx *trxs.TrxContext
	err   xerrors.XError
}

var (
	cases []*Case
)

func init() {

	bz, err := govRuleProposal.Encode()
	if err != nil {
		panic(err)
	}
	cases = []*Case{
		{txctx: makeProposalTrxCtx(stakerHandlerHelper.PickAddress(rand.Int()), libs.ZeroAddress(), big.NewInt(10), bz), err: nil},
		{txctx: makeProposalTrxCtx(libs.RandAddress(), libs.ZeroAddress(), big.NewInt(10), bz), err: xerrors.New("error")},
	}
}

func TestAddProposal(t *testing.T) {
	successes := 0
	for _, c := range cases {
		err := govCtrler.Validate(c.txctx)
		if c.err == nil {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			continue
		}

		err = govCtrler.Apply(c.txctx)
		require.NoError(t, err)
		successes++
	}
	require.Equal(t, successes, len(govCtrler.allProposals))
	require.Equal(t, successes, len(govCtrler.updatedProposals))

}

func TestSaveProposal(t *testing.T) {
	successPorposals := govCtrler.updatedProposals

	_, _, err := govCtrler.Commit()
	require.NoError(t, err)

	err = govCtrler.Close()
	require.NoError(t, err)

	govCtrler2, err := NewGovCtrler(DBDIR, tmlog.NewNopLogger())
	require.NoError(t, err)

	require.Equal(t, len(successPorposals), len(govCtrler2.allProposals))
	require.Equal(t, 0, len(govCtrler2.updatedProposals))

	for _, p0 := range successPorposals {
		p1 := govCtrler2.allProposals[types.HexBytes(p0.ID()).Array32()]
		require.Equal(t, p0, p1)
	}
}
