package gov

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	trxCtxProposal        *ctrlertypes.TrxContext
	voteTestCases1        []*Case
	voteTestCases2        []*Case
	testFlagAlreadyFrozen = false
	baseFee               = uint256.NewInt(1_000_000_000_000_000)
)

func init() {
	bzOpt, err := json.Marshal(govRule1)
	if err != nil {
		panic(err)
	}
	txProposal := web3.NewTrxProposal(
		stakeHelper.PickAddress(1), types.ZeroAddress(), 1, baseFee,
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // wrong min fee
	trxCtxProposal = makeTrxCtx(txProposal, 1, true)
	if xerr := runTrx(trxCtxProposal); xerr != nil {
		panic(xerr)
	}
	if _, _, xerr := govCtrler.Commit(); xerr != nil {
		panic(xerr)
	}

	// no error
	tx0 := web3.NewTrxVoting(stakeHelper.PickAddress(0), types.ZeroAddress(), 1, baseFee,
		trxCtxProposal.TxHash, 0)

	// no right
	tx1 := web3.NewTrxVoting(stakeHelper.PickAddress(stakeHelper.valCnt), types.ZeroAddress(), 1, baseFee,
		trxCtxProposal.TxHash, 0)

	// invalid payload params : wrong choice
	tx2 := web3.NewTrxVoting(stakeHelper.PickAddress(0), types.ZeroAddress(), 1, baseFee,
		trxCtxProposal.TxHash, 1)
	// invalid payload params : wrong choice
	tx3 := web3.NewTrxVoting(stakeHelper.PickAddress(0), types.ZeroAddress(), 1, baseFee,
		trxCtxProposal.TxHash, -1)
	// not found result
	tx4 := web3.NewTrxVoting(stakeHelper.PickAddress(0), types.ZeroAddress(), 1, baseFee,
		bytes.RandBytes(32), 0)

	// test cases #1
	voteTestCases1 = []*Case{
		{txctx: makeTrxCtx(tx0, 1, true), err: xerrors.ErrNotVotingPeriod},           // not voting period
		{txctx: makeTrxCtx(tx0, 10+259200+1, true), err: xerrors.ErrNotVotingPeriod}, // not voting period
		{txctx: makeTrxCtx(tx1, 10, true), err: xerrors.ErrNoRight},                  // no right
		{txctx: makeTrxCtx(tx2, 10, true), err: xerrors.ErrInvalidTrxPayloadParams},  // not found result
		{txctx: makeTrxCtx(tx3, 10, true), err: xerrors.ErrInvalidTrxPayloadParams},  // not found result
		{txctx: makeTrxCtx(tx4, 10, true), err: xerrors.ErrNotFoundResult},           // not found result
		{txctx: makeTrxCtx(tx0, 10, true), err: nil},                                 // success
	}

	// txs of validators except stakeHelper.delegatees[0]
	var txs []*ctrlertypes.Trx
	for i := 1; i < stakeHelper.valCnt; i++ {
		addr := stakeHelper.PickAddress(i)
		choice := int32(0)
		//rn := int(bytes.RandInt63n(int64(len(stakeHelper.delegatees))))
		//if rn%3 == 0 {
		//	choice = 1
		//}
		tx := web3.NewTrxVoting(addr, types.ZeroAddress(), 1, baseFee,
			trxCtxProposal.TxHash, choice)
		txs = append(txs, tx)
	}

	// test cases #2 - all success case
	for i, tx := range txs {
		voteTestCases2 = append(voteTestCases2, &Case{
			txctx: makeTrxCtx(tx, int64(10+i), true),
			err:   nil,
		})
	}
}

func TestVoting(t *testing.T) {
	votedPowers := int64(0)
	for i, c := range voteTestCases1 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)

		if xerr == nil {
			votedPowers += stakeHelper.TotalPowerOf(c.txctx.Tx.From)
		}
	}

	_, _, xerr := govCtrler.Commit()
	require.NoError(t, xerr)

	prop, xerr := govCtrler.ReadProposal(trxCtxProposal.TxHash)
	require.NoError(t, xerr)

	sumVotedPowers := int64(0)
	for i, c := range voteTestCases1 {
		if c.err == nil {
			power := stakeHelper.TotalPowerOf(c.txctx.Tx.From)
			require.Equal(t, power, prop.Options[0].Votes(), "index", i)
			sumVotedPowers += prop.Options[0].Votes()
		}
	}

	require.Equal(t, votedPowers, sumVotedPowers)
}

func TestMajority(t *testing.T) {
	prop, xerr := govCtrler.ReadProposal(trxCtxProposal.TxHash)
	require.NoError(t, xerr)
	require.NotNil(t, prop)

	opt := prop.UpdateMajorOption()
	require.Nil(t, opt)

	votedPowers := prop.Options[0].Votes()
	for i, c := range voteTestCases2 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)

		_, _, xerr = govCtrler.Commit()
		require.NoError(t, xerr)

		prop, xerr := govCtrler.ReadProposal(trxCtxProposal.TxHash)
		require.NoError(t, xerr)
		require.NotNil(t, prop)

		votedPowers += stakeHelper.TotalPowerOf(c.txctx.Tx.From)
		if votedPowers >= prop.MajorityPower {
			opt := prop.UpdateMajorOption()
			require.NotNil(t, opt, votedPowers, prop.MajorityPower)
			require.EqualValues(t, prop.MajorOption, opt)
			require.Equal(t, votedPowers, opt.Votes())
		} else {
			opt := prop.UpdateMajorOption()
			require.Nil(t, opt)
		}
	}

	//
	// duplicated voting
	// its votes MUST not changed
	for i, c := range voteTestCases2 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)

		_, _, xerr = govCtrler.Commit()
		require.NoError(t, xerr)

		prop, xerr := govCtrler.ReadProposal(trxCtxProposal.TxHash)
		require.NoError(t, xerr)
		require.NotNil(t, prop)

		opt := prop.UpdateMajorOption()
		require.NotNil(t, opt)
		require.EqualValues(t, prop.MajorOption, opt)
		require.Equal(t, votedPowers, opt.Votes())
	}
}

func TestFreezingProposal(t *testing.T) {
	// make the proposal majority
	for i, c := range voteTestCases2 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)
	}
	_, _, xerr := govCtrler.Commit()
	require.NoError(t, xerr)

	prop, xerr := govCtrler.ReadProposal(trxCtxProposal.TxHash)
	require.NoError(t, xerr)

	//
	// not changed
	bctx := &ctrlertypes.BlockContext{}
	bctx.SetHeight(prop.EndVotingHeight)
	_, xerr = govCtrler.EndBlock(bctx)
	require.NoError(t, xerr)

	_, _, xerr = govCtrler.Commit()
	require.NoError(t, xerr)
	prop, xerr = govCtrler.ReadProposal(trxCtxProposal.TxHash)
	require.NoError(t, xerr)

	//
	// freezing the proposal
	bctx = &ctrlertypes.BlockContext{}
	bctx.SetHeight(prop.EndVotingHeight + 1)
	_, xerr = govCtrler.EndBlock(bctx)
	require.NoError(t, xerr)

	_, _, xerr = govCtrler.Commit()
	require.NoError(t, xerr)
	_, xerr = govCtrler.ReadProposal(trxCtxProposal.TxHash)
	require.Equal(t, xerrors.ErrNotFoundProposal, xerr)
	frozenProp, xerr := govCtrler.frozenLedger.Get(trxCtxProposal.TxHash.Array32())
	require.NoError(t, xerr)
	require.NotNil(t, frozenProp.MajorOption)
	// prop.MajorOption is nil, so...
	prop.MajorOption = frozenProp.MajorOption
	require.Equal(t, prop, frozenProp)

	testFlagAlreadyFrozen = true
}

func TestApplyingProposal(t *testing.T) {
	oriGovRule := govCtrler.GovRule
	require.Equal(t, ctrlertypes.DefaultGovRule(), &oriGovRule)

	txProposalPayload, ok := trxCtxProposal.Tx.Payload.(*ctrlertypes.TrxPayloadProposal)
	require.True(t, ok)

	if testFlagAlreadyFrozen == false {
		// make the proposal majority
		for i, c := range voteTestCases2 {
			xerr := runCase(c)
			require.Equal(t, c.err, xerr, "index", i)
		}
		_, _, xerr := govCtrler.Commit()
		require.NoError(t, xerr)

		// freezing the proposal
		bctx := &ctrlertypes.BlockContext{}
		bctx.SetHeight(txProposalPayload.StartVotingHeight + txProposalPayload.VotingPeriodBlocks + 1)
		_, xerr = govCtrler.EndBlock(bctx)
		require.NoError(t, xerr)
		_, _, xerr = govCtrler.Commit()
		require.NoError(t, xerr)
	}

	//
	// not changed
	runHeight := txProposalPayload.StartVotingHeight + txProposalPayload.VotingPeriodBlocks + govCtrler.LazyApplyingBlocks() - 1
	bctx := &ctrlertypes.BlockContext{}
	bctx.SetHeight(runHeight)
	_, xerr := govCtrler.EndBlock(bctx)
	require.NoError(t, xerr)
	_, _, xerr = govCtrler.Commit()
	require.NoError(t, xerr)
	frozenProp, xerr := govCtrler.frozenLedger.Get(trxCtxProposal.TxHash.Array32())
	require.NoError(t, xerr)
	require.NotNil(t, frozenProp)

	//
	// apply new gov rule
	runHeight = txProposalPayload.StartVotingHeight + txProposalPayload.VotingPeriodBlocks + govCtrler.LazyApplyingBlocks()
	bctx = &ctrlertypes.BlockContext{}
	bctx.SetHeight(runHeight)
	_, xerr = govCtrler.EndBlock(bctx)
	require.NoError(t, xerr)
	require.NotNil(t, govCtrler.newGovRule)

	_, _, xerr = govCtrler.Commit()
	require.NoError(t, xerr)
	frozenProp, xerr = govCtrler.frozenLedger.Get(trxCtxProposal.TxHash.Array32())
	require.Equal(t, xerrors.ErrNotFoundResult, xerr)
	require.Nil(t, frozenProp)

	require.NotEqual(t, oriGovRule, govCtrler.GovRule)
	require.Equal(t, govRule1, &govCtrler.GovRule)
}
