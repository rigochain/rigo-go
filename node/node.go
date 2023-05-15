package node

import (
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/rpc"
	"github.com/rigochain/rigo-go/types/crypto"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmp2p "github.com/tendermint/tendermint/p2p"
)

type Provider func(*cfg.Config, []byte, tmlog.Logger) (*tmnode.Node, error)

// NewRigoNode returns a Tendermint node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func NewRigoNode(config *cfg.Config, s []byte, logger tmlog.Logger) (*tmnode.Node, error) {
	nodeKey, err := tmp2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load or gen rigo key %s: %w", config.NodeKeyFile(), err)
	}

	rpcOption := func(node *tmnode.Node) {
		rpc.AddRoutes()

		//
		//_ := tmrpccore.Environment{
		//	ProxyAppQuery:   node.ProxyApp().Query(),
		//	ProxyAppMempool: node.ProxyApp().Mempool(),
		//
		//	StateStore:     nil, //n.stateStore,
		//	BlockStore:     node.BlockStore(),
		//	EvidencePool:   node.EvidencePool(),
		//	ConsensusState: node.ConsensusState(),
		//	P2PPeers:       node.Switch(),
		//	P2PTransport:   node,
		//
		//	PubKey:           nil, //,
		//	GenDoc:           node.GenesisDoc(),
		//	TxIndexer:        nil, //n.txIndexer,
		//	BlockIndexer:     nil, //n.blockIndexer,
		//	ConsensusReactor: node.ConsensusReactor(),
		//	EventBus:         node.EventBus(),
		//	Mempool:          node.Mempool(),
		//
		//	Logger: node.Logger.With("module", "rpc"),
		//
		//	Config: *node.Config().RPC,
		//}
	}

	return tmnode.NewNode(config.Config,
		crypto.LoadOrGenSFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), s),
		nodeKey,
		NewRigoLocalClientCreator(NewRigoApp(config, logger.With("module", "rigo"))), //proxy.NewLocalClientCreator(node.NewRigoApp(config.DBDir(), logger)), //proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		tmnode.DefaultGenesisDocProviderFunc(config.Config),
		tmnode.DefaultDBProvider,
		tmnode.DefaultMetricsProvider(config.Instrumentation),
		logger,
		rpcOption,
	)
}
