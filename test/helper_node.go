package test

import (
	"errors"
	"fmt"
	"github.com/containerd/continuity/fs"
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

type PeerMock struct {
	ChainID string
	Config  *cfg.Config
	nd      *tmnode.Node
	RPCURL  string
	WSEnd   string

	Pass []byte
}

func NewPeerMock(chain, id string, p2pPort, rpcPort int, logLevel string) *PeerMock {
	config := cfg.DefaultConfig()
	config.LogLevel = logLevel
	config.SetRoot(filepath.Join(os.TempDir(), "rigo_test_"+id))
	tmcfg.EnsureRoot(config.RootDir)

	config.P2P.AllowDuplicateIP = true
	config.P2P.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", p2pPort)
	config.RPC.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", rpcPort)
	if err := config.ValidateBasic(); err != nil {
		panic(fmt.Errorf("error in rootConfig file: %v", err))
	}

	return &PeerMock{
		ChainID: chain,
		Config:  config,
		RPCURL:  fmt.Sprintf("http://localhost:%d", rpcPort),
		WSEnd:   fmt.Sprintf("ws://localhost:%d/websocket", rpcPort),
		Pass:    []byte("1111"),
	}
}

func (peer *PeerMock) CopyGenesisFrom(source string) error {
	return fs.CopyFile(
		peer.Config.GenesisFile(),
		source)
}

func (peer *PeerMock) IDAddress() (string, error) {
	if peer.nd == nil {
		return "", errors.New("not created node")
	}
	ni := peer.nd.NodeInfo()
	na, _ := ni.NetAddress()
	return fmt.Sprintf("%s@127.0.0.1:%d", ni.ID(), na.Port), nil
}

func (peer *PeerMock) SetPeers(other *PeerMock) {
	peer.Config.P2P.PersistentPeers, _ = other.IDAddress()
	fmt.Println("SetPeers", peer.Config.P2P.PersistentPeers)
}

func (peer *PeerMock) SetPass(pass []byte) {
	peer.Pass = pass
}

func (peer *PeerMock) Init() error {
	return commands.InitFilesWith(peer.ChainID, peer.Config, peer.Pass)
}

func (peer *PeerMock) Start() error {
	logger := tmlog.NewNopLogger()

	logger = tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	if peer.Config.LogFormat == "json" {
		logger = tmlog.NewTMJSONLogger(tmlog.NewSyncWriter(os.Stdout))
	}
	logger, err := tmflags.ParseLogLevel(peer.Config.LogLevel, logger, tmcfg.DefaultLogLevel)

	peer.nd, err = node.NewRigoNode(peer.Config, peer.Pass, logger)
	if err != nil {
		return fmt.Errorf("failed to create rigo: %w", err)
	}

	err = peer.nd.Start()
	if err != nil {
		return fmt.Errorf("failed to start rigo: %w", err)
	}
	return nil
}

func (peer *PeerMock) Stop() {
	if peer.nd.IsRunning() {
		if err := peer.nd.ProxyApp().Stop(); err != nil {
			panic(fmt.Errorf("unable to stop the rigo proxy app: %v", err))
		}
		if err := peer.nd.Stop(); err != nil {
			panic(fmt.Errorf("unable to stop the rigo node: %v", err))
		}
	}
}

func (peer *PeerMock) WalletPath() string {
	return filepath.Join(peer.Config.RootDir, "walkeys")
}

func (peer *PeerMock) PrivValKeyPath() string {
	return peer.Config.PrivValidatorKeyFile()
}
