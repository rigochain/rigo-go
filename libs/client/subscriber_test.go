package client

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSubscriber(t *testing.T) {
	sub, err := NewSubscriber("ws://localhost:26657/websocket") //NewSubscriber("ws://192.168.252.60:26657/websocket")
	require.NoError(t, err)

	success := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	err = sub.Watch("tm.event='NewBlockHeader'", func(sub *Subscriber, msg []byte) {
		success = len(msg) > 10
		wg.Done()
	})
	require.NoError(t, err)

	wg.Wait()

	require.True(t, success)
}
