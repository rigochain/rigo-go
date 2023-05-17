package gov

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"math/rand"
	"testing"
	"time"
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
		stakeHelper.PickAddress(1), types.ZeroAddress(), 1, uint256.NewInt(5), // insufficient fee
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt)

	tx1 := web3.NewTrxProposal( // no right
		stakeHelper.PickAddress(stakeHelper.valCnt+1), types.ZeroAddress(), 1, uint256.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt)

	tx3 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, uint256.NewInt(10),
		"test govrule proposal", 10, 159200, proposal.PROPOSAL_GOVRULE, bzOpt) // wrong period

	tx4 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, uint256.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // it will be used to test wrong start height

	tx5 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, uint256.NewInt(10),
		"test govrule proposal", 10, 259200, proposal.PROPOSAL_GOVRULE, bzOpt) // all right

	cases1 = []*Case{
		{txctx: makeTrxCtx(tx0, 1, true), err: xerrors.ErrInsufficientFee}, // wrong min fee
		{txctx: makeTrxCtx(tx1, 1, true), err: xerrors.ErrNoRight},
		{txctx: makeTrxCtx(tx3, 1, true), err: xerrors.ErrInvalidTrxPayloadParams},  // wrong period
		{txctx: makeTrxCtx(tx4, 20, true), err: xerrors.ErrInvalidTrxPayloadParams}, // wrong start height
		{txctx: makeTrxCtx(tx5, 1, true), err: nil},                                 // success
	}

	tx6 := web3.NewTrxProposal(
		stakeHelper.PickAddress(stakeHelper.valCnt-1), types.ZeroAddress(), 1, uint256.NewInt(10),
		"test govrule proposal2", 11, 259200, proposal.PROPOSAL_GOVRULE, bzOpt)
	cases2 = []*Case{
		// the tx6 will be submitted two times.
		// the first must success but the second must fail.
		{txctx: makeTrxCtx(tx6, 1, true), err: nil},
	}
}

func TestAddProposal(t *testing.T) {
	props0, _ := govCtrler.RealAllProposals()
	require.NotNil(t, props0)

	for i, c := range cases1 {
		xerr := runCase(c)
		require.Equal(t, c.err, xerr, "index", i)
	}

	props1, _ := govCtrler.RealAllProposals()
	require.NotNil(t, props1)
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

func TestPunish(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	valAddr := stakeHelper.PickAddress(rand.Intn(stakeHelper.valCnt))

	// original proposals and voter's power
	props, err := govCtrler.RealAllProposals()
	require.NoError(t, err)

	type propSnapshot struct {
		proposal      *proposal.GovProposal
		punishedVoter *proposal.Voter
	}
	var props0 []propSnapshot
	for _, prop := range props {
		v, ok := prop.Voters[valAddr.String()]
		if ok {
			props0 = append(props0,
				propSnapshot{
					proposal:      prop,
					punishedVoter: v,
				})
		}
	}

	slashed, err := govCtrler.DoPunish(&abcitypes.Evidence{
		Validator: abcitypes.Validator{
			Address: valAddr,
			Power:   stakeHelper.PowerOf(valAddr),
		},
	})
	require.NoError(t, err)

	// commit
	_, _, err = govCtrler.Commit()
	require.NoError(t, err)

	// proposals and voter's power after punishing
	props, err = govCtrler.RealAllProposals()
	require.NoError(t, err)

	var props1 []propSnapshot
	for _, prop := range props {
		v, ok := prop.Voters[valAddr.String()]
		if ok {
			props1 = append(props1,
				propSnapshot{
					proposal:      prop,
					punishedVoter: v,
				})
		}
	}

	// check punishment result
	require.Equal(t, len(props0), len(props1))
	power0, power1 := int64(0), int64(0)
	for i, prop0 := range props0 {
		require.Equal(t, prop0.proposal.SumVotingPowers(), prop0.proposal.TotalVotingPower)
		require.Equal(t, props1[i].proposal.SumVotingPowers(), props1[i].proposal.TotalVotingPower)
		power0 += prop0.proposal.TotalVotingPower
		power1 += props1[i].proposal.TotalVotingPower
	}

	require.Equal(t, power0-slashed, power1)
}
