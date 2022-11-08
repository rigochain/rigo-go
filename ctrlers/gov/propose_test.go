package gov

import (
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"testing"
)

func TestGovCtrler_ProposalSaving(t *testing.T) {
	txctx := makeTrxCtx()
	require.NotNil(t, txctx)

	err := govCtrler.Apply(txctx)
	require.NoError(t, err)

	originProposals := govCtrler.allProposals
	require.Equal(t, 1, len(govCtrler.allProposals))
	require.Equal(t, 1, len(govCtrler.updatedProposals))

	_, _, err = govCtrler.Commit()
	require.NoError(t, err)

	err = govCtrler.Close()
	require.NoError(t, err)

	govCtrler, err = NewGovCtrler(DBDIR, tmlog.NewNopLogger())
	require.NoError(t, err)

	require.Equal(t, 1, len(govCtrler.allProposals))
	require.Equal(t, 0, len(govCtrler.updatedProposals))

	p0 := originProposals[txctx.TxHash.Array32()]
	p1 := govCtrler.allProposals[txctx.TxHash.Array32()]

	require.Equal(t, p0, p1)
}
