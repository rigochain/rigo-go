package gov

import (
	"encoding/json"
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var (
	newTrxContext *ctrlertypes.TrxContext
	voteCase      []*Case
)

func init() {
	bzOpt, err := json.Marshal(govParams3)
	if err != nil {
		panic(err)
	}
	newTrx := web3.NewTrxProposal(
		stakeHelper.PickAddress(1), types.ZeroAddress(), 1, defMinGas, defGasPrice,
		"test improving governance parameters proposal", 15, 259200, proposal.PROPOSAL_GOVPARAMS, bzOpt)
	newTrxContext = makeTrxCtx(newTrx, 1, true)
	if xerr := runTrx(newTrxContext); xerr != nil {
		panic(xerr)
	}
	if _, _, xerr := govCtrler.Commit(); xerr != nil {
		panic(xerr)
	}

	var txs []*ctrlertypes.Trx
	for i := 1; i < stakeHelper.valCnt; i++ {
		addr := stakeHelper.PickAddress(i)
		choice := int32(0)
		tx := web3.NewTrxVoting(addr, types.ZeroAddress(), 1, defMinGas, defGasPrice,
			newTrxContext.TxHash, choice)
		txs = append(txs, tx)
	}

	for i, tx := range txs {
		voteCase = append(voteCase, &Case{
			txctx: makeTrxCtx(tx, int64(10+i), true),
			err:   nil,
		})
	}
}

func TestMergeGovParams(t *testing.T) {
	oriParams := govCtrler.GovParams
	newParams := ctrlertypes.DefaultGovParams()

	ctrlertypes.MergeGovParams(&oriParams, newParams)
	if !reflect.DeepEqual(newParams, ctrlertypes.DefaultGovParams()) {
		t.Errorf("unexpected GovParams: %v", newParams)
	}
}

func TestApplyMergeProposal(t *testing.T) {
	for _, c := range voteCase {
		runCase(c)
	}
	govCtrler.Commit()

	blockContext := &ctrlertypes.BlockContext{}
	txProposalPayload := newTrxContext.Tx.Payload.(*ctrlertypes.TrxPayloadProposal)
	blockContext.SetHeight(txProposalPayload.StartVotingHeight + txProposalPayload.VotingPeriodBlocks + 1)
	govCtrler.EndBlock(blockContext)
	govCtrler.Commit()

	txProposalPayload, ok := newTrxContext.Tx.Payload.(*ctrlertypes.TrxPayloadProposal)
	require.True(t, ok)

	runHeight := txProposalPayload.StartVotingHeight + txProposalPayload.VotingPeriodBlocks + govCtrler.LazyApplyingBlocks()
	blockContext = &ctrlertypes.BlockContext{}
	blockContext.SetHeight(runHeight)
	_, xerr := govCtrler.EndBlock(blockContext)
	require.NoError(t, xerr)
	_, _, xerr = govCtrler.Commit()
	require.NoError(t, xerr)
}
