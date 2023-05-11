package test

import (
	"fmt"
	"github.com/rigochain/rigo-go/cmd/commands"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/node"
	tmcfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	"os"
	"path/filepath"
)

var (
	testChainID = "rigo_unit_test_net"
	testConfig  *cfg.Config
	nd          *tmnode.Node
	rpcURL      = "http://localhost:26657"
	wsEndpoint  = "ws://localhost:26657/websocket"

	TESTPASS = []byte("1111")
)

func init() {
	testConfig = cfg.DefaultConfig()
	testConfig.LogLevel = "none"
	testConfig.SetRoot(filepath.Join(os.TempDir(), "rigo_test"))
	tmcfg.EnsureRoot(testConfig.RootDir)
	fmt.Println("root directory", testConfig.RootDir)
	testConfig.RPC.ListenAddress = "tcp://localhost:36657"
	if err := testConfig.ValidateBasic(); err != nil {
		panic(fmt.Errorf("error in rootConfig file: %v", err))
	}

	rpcURL = "http://localhost:36657"
	wsEndpoint = "ws://localhost:36657/websocket"
}

func initNode() error {
	return commands.InitFilesWith(testChainID, testConfig, TESTPASS)
}

func runNode() error {
	var err error

	logger := tmlog.NewNopLogger()

	logger = tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	if testConfig.LogFormat == "json" {
		logger = tmlog.NewTMJSONLogger(tmlog.NewSyncWriter(os.Stdout))
	}
	logger, err = tmflags.ParseLogLevel(testConfig.LogLevel, logger, tmcfg.DefaultLogLevel)

	nd, err = node.NewRigoNode(testConfig, TESTPASS, logger)
	if err != nil {
		return fmt.Errorf("failed to create rigo: %w", err)
	}

	err = nd.Start()
	if err != nil {
		return fmt.Errorf("failed to start rigo: %w", err)
	}

	return nil
}

func stopNode() {
	if nd.IsRunning() {
		if err := nd.ProxyApp().Stop(); err != nil {
			panic(fmt.Errorf("unable to stop the rigo proxy app: %v", err))
		}
		if err := nd.Stop(); err != nil {
			panic(fmt.Errorf("unable to stop the rigo node: %v", err))
		}
	}
}

func walletPath() string {
	return filepath.Join(testConfig.Config.RootDir, "walkeys")
}

func privValKeyPath() string {
	return testConfig.PrivValidatorKeyFile()
}
