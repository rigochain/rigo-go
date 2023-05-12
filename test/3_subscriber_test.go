package test

import (
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"sync"
	"testing"
)

func TestSubscriber(t *testing.T) {
	sub, err := web3.NewSubscriber(wsEndpoint)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	err = sub.Start("tm.event='NewBlock'", func(sub *web3.Subscriber, result []byte) {
		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)
		_, ok := event.Data.(types.EventDataNewBlock)
		require.True(t, ok)

		wg.Done()
	})
	require.NoError(t, err)
	wg.Wait()

	sub.Stop()
}
