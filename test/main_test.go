package test

import (
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	tmcfg "github.com/tendermint/tendermint/config"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"os"
	"strconv"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	test_on_internal_node(m)
	//test_on_external_node(m)
}

var (
	peers []*PeerMock
)

func test_on_internal_node(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		for i := 0; i < 10; i++ {
			ll := "none"
			//if i == 9 {
			//	ll = "rigo:debug,*:error"
			//}
			_peer := NewPeerMock("rigo_test_chain", strconv.FormatInt(int64(i), 10), 26656+i, 36657+i, ll)
			if err := _peer.Init(); err != nil {
				panic(err)
			}

			if i > 0 {
				prevPeer := peers[i-1]
				if err := _peer.CopyGenesisFrom(prevPeer.Config.GenesisFile()); err != nil {
					panic(err)
				}
				_peer.SetPeers(prevPeer)
			}

			if err := _peer.Start(); err != nil {
				panic(err)
			}

			peers = append(peers, _peer)
		}
		wg.Done()
	}()
	wg.Wait()

	rpcNode = peers[0]
	subWg, err := waitEvent("tm.event='NewBlock'", func(event *coretypes.ResultEvent, err error) bool {
		return true
	})
	if err != nil {
		panic(err)
	}
	subWg.Wait()

	prepareTest(peers[0])

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
	rpcNode = peer

	prepareTest(peer)

	exitCode := m.Run()

	os.Exit(exitCode)
}
