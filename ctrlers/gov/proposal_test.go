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
	"math/big"
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
	bzOpt, err := json.Marshal(govRule0)
	if err != nil {
		panic(err)
	}

	tx0 := web3.NewTrxProposal(
		stakeHelper.PickAddress(1), types.ZeroAddress(), 1, big.NewInt(5),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // wrong min fee

	tx1 := web3.NewTrxProposal( // no right
		stakeHelper.PickAddress(stakeHelper.valCnt+1), types.ZeroAddress(), 1, big.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt)

	// Not testable,
	// TrxContext can not be made because tx.Encode()/Decode() returns error
	//
	//h, err := tx1.Hash()
	//if err != nil {
	//	panic(err)
	//}
	//tx2 := web3.NewTrxVoting(
	//	stakeHelper.PickAddress(1), types.ZeroAddress(), 1, big.NewInt(10),
	//	h, 0)
	//tx2.Type = ctrlertypes.TRX_PROPOSAL // wrong payload type

	// Not testable,
	// In ValidateTrx/ExecuteTrx, option is not decoded and checked
	//
	//tx3 := web3.NewTrxProposal(
	//	stakeHelper.PickAddress(stakeHelper.valCnt+1), types.ZeroAddress(), 1, big.NewInt(10),
	//	"test govrule proposal", 10, 259200, proposal.PROPOSAL_COMMON, bz) // wrong option type

	tx4 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, big.NewInt(10),
		"test govrule proposal", 10, 159200, proposal.PROPOSAL_GOVRULE, bzOpt) // wrong period

	tx5 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, big.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // it will be used to test wrong start height

	tx6 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, big.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // all right

	cases1 = []*Case{
		{txctx: makeTrxCtx(tx0, 1, true), err: xerrors.ErrInsufficientFee}, // wrong min fee
		{txctx: makeTrxCtx(tx1, 1, true), err: xerrors.ErrNoRight},
		{txctx: makeTrxCtx(tx4, 1, true), err: xerrors.ErrInvalidTrxPayloadParams},  // wrong period
		{txctx: makeTrxCtx(tx5, 20, true), err: xerrors.ErrInvalidTrxPayloadParams}, // wrong start height
		{txctx: makeTrxCtx(tx6, 1, true), err: nil},                                 // success
	}

	tx7 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, big.NewInt(10),
		"test govrule proposal2", 11, 259200, proposal.PROPOSAL_GOVRULE, bzOpt)
	cases2 = []*Case{
		{txctx: makeTrxCtx(tx7, 1, true), err: nil}, // first success, second fail
	}
}

func TestAddProposal(t *testing.T) {
	for i, c := range cases1 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)
	}
}

func TestCommitProposal(t *testing.T) {
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

func runCase(c *Case) xerrors.XError {
	return runTrx(c.txctx)
}

func runTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	if xerr := govCtrler.ValidateTrx(ctx); xerr != nil {
		return xerr
	}
	if xerr := govCtrler.ExecuteTrx(ctx); xerr != nil {
		return xerr
	}
	return nil
}
