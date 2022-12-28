package node

import (
	"fmt"
	cfg "github.com/kysee/arcanus/cmd/config"
	"github.com/kysee/arcanus/rpc"
	"github.com/kysee/arcanus/types/crypto"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
)

type Provider func(*cfg.Config, []byte, tmlog.Logger) (*tmnode.Node, error)

// NewArcanusNode returns a Tendermint node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func NewArcanusNode(config *cfg.Config, s []byte, logger tmlog.Logger) (*tmnode.Node, error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load or gen arcanus key %s: %w", config.NodeKeyFile(), err)
	}

	rpcOption := func(node *tmnode.Node) {
		rpc.AddRoutes()
	}

	return tmnode.NewNode(config.Config,
		crypto.LoadOrGenSFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), s),
		nodeKey,
		NewArcanusLocalClientCreator(NewArcanusApp(config, logger.With("module", "arcanus"))), //proxy.NewLocalClientCreator(node.NewArcanusApp(config.DBDir(), logger)), //proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		tmnode.DefaultGenesisDocProviderFunc(config.Config),
		tmnode.DefaultDBProvider,
		tmnode.DefaultMetricsProvider(config.Instrumentation),
		logger,
		rpcOption,
	)
}
