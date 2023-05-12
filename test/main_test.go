package test

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	//test_on_internal_node(m)
	test_on_external_node(m)
}

func test_on_internal_node(m *testing.M) {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err = initNode(); err != nil {
			panic(err)
		}
		if err = runNode(); err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
		wg.Done()
	}()
	wg.Wait()

	prepareTest()

	exitCode := m.Run()

	stopNode()

	os.RemoveAll(testConfig.RootDir)
	os.Exit(exitCode)
}

func test_on_external_node(m *testing.M) {
	// node to be executed externally
	rpcURL = "http://localhost:26657"
	wsEndpoint = "ws://localhost:26657/websocket"
	testConfig.RootDir = "/Users/kylekwon/rigo_localnet_0"
	TESTPASS = []byte("1")

	prepareTest()

	exitCode := m.Run()

	os.Exit(exitCode)
}
