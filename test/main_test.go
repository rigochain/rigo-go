package test

import (
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	tmcfg "github.com/tendermint/tendermint/config"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"os"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	test_on_internal_node(m)
	//test_on_external_node(m)
}

func test_on_internal_node(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		runPeers(3)

		wg.Done()
	}()
	wg.Wait()

	defaultRpcNode = peers[len(peers)-1]
	subWg, err := waitEvent("tm.event='NewBlock'", func(event *coretypes.ResultEvent, err error) bool {
		return true
	})
	if err != nil {
		panic(err)
	}
	subWg.Wait()

	prepareTest(peers[:1]) // peers[0] is active validator node

	exitCode := m.Run()

	for _, p := range peers {
		p.Stop()
		os.RemoveAll(p.Config.RootDir)
	}

	os.Exit(exitCode)
}

func test_on_external_node(m *testing.M) {
	//// node to be executed externally
	//rpcURL = "http://localhost:26657"
	//wsEndpoint = "ws://localhost:26657/websocket"
	//config.RootDir = "/Users/kylekwon/rigo_localnet_0"
	//TESTPASS = []byte("1")

	config := cfg.DefaultConfig()
	config.LogLevel = ""
	config.SetRoot("/Users/kysee/rigo_localnet_0")
	tmcfg.EnsureRoot(config.RootDir)
	if err := config.ValidateBasic(); err != nil {
		panic(fmt.Errorf("error in rootConfig file: %v", err))
	}

	peer := &PeerMock{
		ChainID: "rigo_test_external_chain",
		Config:  config,
		RPCURL:  "http://localhost:26657",
		WSEnd:   "ws://localhost:26657/websocket",
		Pass:    []byte("1"),
	}
	defaultRpcNode = peer

	prepareTest([]*PeerMock{peer})

	exitCode := m.Run()

	os.Exit(exitCode)
}
