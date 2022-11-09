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

var cases []*Case

func init() {
	cases = []*Case{
		{txctx: makeProposalTrxCtx(stakerCtrler.PickAddress(rand.Int()), libs.ZeroAddress(), big.NewInt(10)), err: nil},
		{txctx: makeProposalTrxCtx(libs.RandAddress(), libs.ZeroAddress(), big.NewInt(10)), err: xerrors.New("error")},
	}
}

func TestProposalAndSaving(t *testing.T) {
	proposed := 0
	for _, c := range cases {
		err := govCtrler.Apply(c.txctx)
		if c.err == nil {
			require.NoError(t, err)
			proposed++
		} else {
			require.Error(t, err)
		}
	}

	originProposals := govCtrler.updatedProposals
	require.Equal(t, proposed, len(govCtrler.allProposals))
	require.Equal(t, proposed, len(govCtrler.updatedProposals))

	_, _, err := govCtrler.Commit()
	require.NoError(t, err)

	err = govCtrler.Close()
	require.NoError(t, err)

	govCtrler, err = NewGovCtrler(DBDIR, tmlog.NewNopLogger())
	require.NoError(t, err)

	require.Equal(t, proposed, len(govCtrler.allProposals))
	require.Equal(t, 0, len(govCtrler.updatedProposals))

	for i := 0; i < proposed; i++ {
		p0 := originProposals[i]
		p1 := govCtrler.allProposals[types.HexBytes(p0.ID()).Array32()]
		require.Equal(t, p0, p1)
	}
}
