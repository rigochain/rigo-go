package test

import (
	"encoding/json"
	"fmt"
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"sync"
	"testing"
)

var (
	proposalHash      bytes.HexBytes
	targetBlockHeight int64
)

func TestProposalAndVoting(t *testing.T) {
	//get validator wallet
	rweb3 := randRigoWeb3()
	validatorWallets := validatorWallets[0]
	require.NoError(t, validatorWallets.SyncAccount(rweb3))
	require.NoError(t, validatorWallets.Unlock(defaultRpcNode.Pass))

	bzOpt, err := json.Marshal(ctrlertypes.Test3GovRule())
	require.NoError(t, err)

	//subscriber
	subWg := sync.WaitGroup{}
	sub, err := web3.NewSubscriber(defaultRpcNode.WSEnd)
	require.NoError(t, err)

	subWg.Add(2)
	propSubErr := sub.Start("tm.event='NewBlock'", func(sub *web3.Subscriber, result []byte) {
		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)
		currentBlockHeight := event.Data.(types.EventDataNewBlock).Block.Height
		if currentBlockHeight > 10 {
			targetBlockHeight = currentBlockHeight + 5
			proposalResult, err2 := validatorWallets.ProposalCommit(defGas, defGasPrice, "proposal test", targetBlockHeight, 259200, proposal.PROPOSAL_GOVRULE, bzOpt, rweb3)
			require.NoError(t, err2)
			require.Equal(t, xerrors.ErrCodeSuccess, proposalResult.CheckTx.Code)
			proposalHash = bytes.HexBytes(proposalResult.Hash)
			sub.Stop()
			subWg.Done()
		}
	})
	require.NoError(t, propSubErr)

	sub2, err := web3.NewSubscriber(defaultRpcNode.WSEnd)
	require.NoError(t, err)

	votingSubErr := sub2.Start("tm.event='NewBlock'", func(sub *web3.Subscriber, result []byte) {
		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)

		votingBlockHeight := event.Data.(types.EventDataNewBlock).Block.Height
		if votingBlockHeight > targetBlockHeight && proposalHash != nil {
			require.NoError(t, validatorWallets.SyncAccount(rweb3))
			votingResult, err := validatorWallets.VotingCommit(defGas, defGasPrice, proposalHash, 0, rweb3)
			require.NoError(t, err)
			require.Equal(t, xerrors.ErrCodeSuccess, votingResult.CheckTx.Code)
			sub2.Stop()
			subWg.Done()
		} else if votingBlockHeight > targetBlockHeight+100 {
			require.NoError(t, fmt.Errorf("no proposal hash for voting"))
			sub2.Stop()
			subWg.Done()
		}
	})
	require.NoError(t, votingSubErr)
	subWg.Wait()
}
