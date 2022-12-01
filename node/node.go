package node

import (
	"fmt"
	arcanus "github.com/kysee/arcanus/ctrlers/state"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/rpc"
	cfg "github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
)

type Provider func(*cfg.Config, []byte, tmlog.Logger) (*tmnode.Node, error)

// XChainNewNode returns a Tendermint node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func XChainNewNode(config *cfg.Config, s []byte, logger tmlog.Logger) (*tmnode.Node, error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load or gen node key %s: %w", config.NodeKeyFile(), err)
	}

	rpcOption := func(node *tmnode.Node) {
		rpc.AddRoutes()
	}

	return tmnode.NewNode(config,
		crypto.LoadOrGenSFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), s),
		nodeKey,
		arcanus.NewArcanusLocalClientCreator(arcanus.NewChainCtrler(config.DBDir(), logger)), //proxy.NewLocalClientCreator(arcanus.NewChainCtrler(config.DBDir(), logger)), //proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		tmnode.DefaultGenesisDocProviderFunc(config),
		tmnode.DefaultDBProvider,
		tmnode.DefaultMetricsProvider(config.Instrumentation),
		logger,
		rpcOption,
	)
}
