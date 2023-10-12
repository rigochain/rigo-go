package gov

import (
	"encoding/json"
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

type Case struct {
	txctx *ctrlertypes.TrxContext
	err   xerrors.XError
}

var (
	cases1 []*Case
	cases2 []*Case
)

func init() {
	bzOpt, err := json.Marshal(govParams0)
	if err != nil {
		panic(err)
	}

	//tx0 := web3.NewTrxProposal(
	//	stakeHelper.PickAddress(1), types.ZeroAddress(), 1, 99_999, defGasPrice, // insufficient fee
	//	"test govparams proposal", 10, 259200, proposal.PROPOSAL_GOVPARAMS, bzOpt)

	tx1 := web3.NewTrxProposal( // no right
		stakeHelper.PickAddress(stakeHelper.valCnt+1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt)

	tx3 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 159200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt) // wrong period

	tx4 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt) // it will be used to test wrong start height

	tx5 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt) // all right

	cases1 = []*Case{
		//{txctx: makeTrxCtx(tx0, 1, true), err: xerrors.ErrInvalidGas}, // wrong min fee
		{txctx: makeTrxCtx(tx1, 1, true), err: xerrors.ErrNoRight},
		{txctx: makeTrxCtx(tx3, 1, true), err: xerrors.ErrInvalidTrxPayloadParams},  // wrong period
		{txctx: makeTrxCtx(tx4, 20, true), err: xerrors.ErrInvalidTrxPayloadParams}, // wrong start height
		{txctx: makeTrxCtx(tx5, 1, true), err: nil},                                 // success

	}

	tx6 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal2", 11, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	cases2 = []*Case{
		// the tx6 will be submitted two times.
		// the first must success but the second must fail.
		{txctx: makeTrxCtx(tx6, 1, true), err: nil},
	}
}

func TestAddProposal(t *testing.T) {
	props0, _ := govCtrler.ReadAllProposals()
	require.NotNil(t, props0)

	for i, c := range cases1 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)
	}

	props1, _ := govCtrler.ReadAllProposals()
	require.NotNil(t, props1)
}

func TestProposalDuplicate(t *testing.T) {
	for i, c := range cases2 {
		require.NoError(t, runCase(c), "index", i)
	}

	_, _, err := govCtrler.Commit()
	require.NoError(t, err)

	for _, c := range cases2 {
		key := ledger.ToLedgerKey(c.txctx.TxHash)
		prop, xerr := govCtrler.proposalLedger.Get(key)
		require.NoError(t, xerr)
		require.NotNil(t, prop)
		require.Equal(t, key, prop.Key())
	}
	for i, c := range cases2 {
		require.Error(t, xerrors.ErrDuplicatedKey, runCase(c), "index", i)
	}
}

func TestOverflowBlockHeight(t *testing.T) {
	bzOpt, err := json.Marshal(govParams0)
	require.NoError(t, err)

	tx := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", math.MaxInt64, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	xerr := runTrx(makeTrxCtx(tx, 1, true))
	require.Error(t, xerr)
	require.Contains(t, xerr.Error(), "overflow occurs")
}

func TestApplyHeight(t *testing.T) {
	bzOpt, err := json.Marshal(govParams0)
	require.NoError(t, err)

	tx0 := web3.NewTrxProposal( // applyHeight : 0
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, 0, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	xerr := runTrx(makeTrxCtx(tx0, 1, true))
	require.NoError(t, xerr)

	tx1 := web3.NewTrxProposal( // applyHeight : start + period + lazyApplyingBlocks
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, ctrlertypes.DefaultGovParams().LazyApplyingBlocks()+259200+10, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	xerr = runTrx(makeTrxCtx(tx1, 1, true))
	require.NoError(t, xerr)

	tx2 := web3.NewTrxProposal( // wrong applyHeight
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, 1, proposal.PROPOSAL_GOVPARAMS, bzOpt)

	xerr = runTrx(makeTrxCtx(tx2, 1, true))
	require.Error(t, xerr)
	require.Contains(t, xerr.Error(), "wrong applyHeight")

	tx3 := web3.NewTrxProposal( // applyHeight : start + period + lazyApplyingBlocks - 1
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, ctrlertypes.DefaultGovParams().LazyApplyingBlocks()+259200+10-1, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	xerr = runTrx(makeTrxCtx(tx3, 1, true))
	require.Error(t, xerr)
	require.Contains(t, xerr.Error(), "wrong applyHeight")

	tx4 := web3.NewTrxProposal( // applyHeight : -518410
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test govparams proposal", 10, 259200, -518410, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	xerr = runTrx(makeTrxCtx(tx4, 1, true))
	require.Error(t, xerr)
	require.Contains(t, xerr.Error(), "wrong applyHeight")

}
