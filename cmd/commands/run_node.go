package commands

import (
	"bytes"
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/libs/sfeeder/client"
	"github.com/rigochain/rigo-go/node"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	tmp2p "github.com/tendermint/tendermint/p2p"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	genesisHash []byte
)

// AddNodeFlags exposes some common configuration options on the command-line
// These are exposed for convenience of commands embedding a node
func AddNodeFlags(cmd *cobra.Command) {
	// bind flags
	cmd.Flags().String("moniker", rootConfig.Moniker, "rigo name")

	// priv val flags
	cmd.Flags().String(
		"priv_validator_laddr",
		rootConfig.PrivValidatorListenAddr,
		"socket address to listen on for connections from external priv_validator process")

	cmd.Flags().StringVar(
		&privValSecret,
		"priv_validator_secret",
		"",
		"passphrase to encrypt and decrypt a private key in priv_validator_key.json",
	)

	cmd.Flags().StringVar(
		&privValSecretFeederAddr,
		"priv_validator_secret_feeder",
		"",
		"socket address to listen on for connections from external priv_validator process")

	// node flags
	cmd.Flags().Bool("fast_sync", rootConfig.FastSyncMode, "fast blockchain syncing")
	cmd.Flags().BytesHexVar(
		&genesisHash,
		"genesis_hash",
		[]byte{},
		"optional SHA-256 hash of the genesis file")
	cmd.Flags().Int64("consensus.double_sign_check_height", rootConfig.Consensus.DoubleSignCheckHeight,
		"how many blocks to look back to check existence of the rigo's "+
			"consensus votes before joining consensus")

	// abci flags
	cmd.Flags().String(
		"proxy_app",
		rootConfig.ProxyApp,
		"proxy app address, or one of: 'kvstore',"+
			" 'persistent_kvstore',"+
			" 'counter',"+
			" 'counter_serial' or 'noop' for local testing.")
	cmd.Flags().String("abci", rootConfig.ABCI, "specify abci transport (socket | grpc)")

	// provider flags
	cmd.Flags().String("rpc.laddr", rootConfig.RPC.ListenAddress, "RPC listen address. Port required")
	cmd.Flags().StringSlice("rpc.cors_allowed_origins", rootConfig.RPC.CORSAllowedOrigins, "")
	cmd.Flags().String(
		"rpc.grpc_laddr",
		rootConfig.RPC.GRPCListenAddress,
		"GRPC listen address (BroadcastTx only). Port required")
	cmd.Flags().Bool("rpc.unsafe", rootConfig.RPC.Unsafe, "enabled unsafe provider methods")
	cmd.Flags().String("rpc.pprof_laddr", rootConfig.RPC.PprofListenAddress, "pprof listen address (https://golang.org/pkg/net/http/pprof)")

	// p2p flags
	cmd.Flags().String(
		"p2p.laddr",
		rootConfig.P2P.ListenAddress,
		"rigo listen address. (0.0.0.0:0 means any interface, any port)")
	cmd.Flags().String("p2p.seeds", rootConfig.P2P.Seeds, "comma-delimited ID@host:port seed nodes")
	cmd.Flags().String("p2p.persistent_peers", rootConfig.P2P.PersistentPeers, "comma-delimited ID@host:port persistent peers")
	cmd.Flags().String("p2p.unconditional_peer_ids",
		rootConfig.P2P.UnconditionalPeerIDs, "comma-delimited IDs of unconditional peers")
	cmd.Flags().Bool("p2p.upnp", rootConfig.P2P.UPNP, "enable/disable UPNP port forwarding")
	cmd.Flags().Bool("p2p.pex", rootConfig.P2P.PexReactor, "enable/disable Peer-Exchange")
	cmd.Flags().Bool("p2p.seed_mode", rootConfig.P2P.SeedMode, "enable/disable seed mode")
	cmd.Flags().String("p2p.private_peer_ids", rootConfig.P2P.PrivatePeerIDs, "comma-delimited private peer IDs")

	// consensus flags
	cmd.Flags().Bool(
		"consensus.create_empty_blocks",
		rootConfig.Consensus.CreateEmptyBlocks,
		"set this to false to only produce blocks when there are txs or when the AppHash changes")
	cmd.Flags().String(
		"consensus.create_empty_blocks_interval",
		rootConfig.Consensus.CreateEmptyBlocksInterval.String(),
		"the possible interval between empty blocks")

	// db flags
	cmd.Flags().String(
		"db_backend",
		rootConfig.DBBackend,
		"database backend: goleveldb | cleveldb | boltdb | rocksdb | badgerdb")
	cmd.Flags().String(
		"db_dir",
		rootConfig.DBPath,
		"database directory")
}

// NewRunNodeCmd returns the command that allows the CLI to start a node.
// It can be used with a custom PrivValidator and in-process ABCI application.
func NewRunNodeCmd(nodeProvider node.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"run"},
		Short:   "Run the rigo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkGenesisHash(rootConfig); err != nil {
				return err
			}

			var s []byte
			if privValSecretFeederAddr != "" {
				wkf, err := crypto.OpenWalletKey(libs.NewFileReader(rootConfig.PrivValidatorKeyFile()))
				if err != nil {
					return err
				}
				nodeKey, err := tmp2p.LoadNodeKey(rootConfig.NodeKeyFile())

				if err != nil {
					return err
				}
				s, err = client.ReadRemoteCredential(string(nodeKey.ID()), privValSecretFeederAddr, wkf.Address)
				if err != nil {
					return err
				}
			} else if privValSecret != "" {
				s = []byte(privValSecret)
				privValSecret = ""
			} else {
				s = libs.ReadCredential(fmt.Sprintf("Passphrase for %v: ", filepath.Base(rootConfig.PrivValidatorKeyFile())))
			}

			n, err := nodeProvider(rootConfig, s, logger)

			libs.ClearCredential(s)

			if err != nil {
				return fmt.Errorf("failed to create rigo: %w", err)
			}

			if err := n.Start(); err != nil {
				return fmt.Errorf("failed to start rigo: %w", err)
			}

			logger.Info("Started rigo", "nodeInfo", n.Switch().NodeInfo())

			// Stop upon receiving SIGTERM or CTRL-C.
			trapSignal(logger, func() {
				if n.IsRunning() {
					if err := n.ProxyApp().Stop(); err != nil {
						logger.Error("unable to stop the rigo proxy app", "error", err)
					}
					if err := n.Stop(); err != nil {
						logger.Error("unable to stop the rigo node", "error", err)
					}
				}
			})

			// Run forever.
			select {}
		},
	}

	AddNodeFlags(cmd)
	return cmd
}

func checkGenesisHash(config *cfg.Config) error {
	if len(genesisHash) == 0 || config.Genesis == "" {
		return nil
	}

	// Calculate SHA-256 hash of the genesis file.
	f, err := os.Open(config.GenesisFile())
	if err != nil {
		return fmt.Errorf("can't open genesis file: %w", err)
	}
	defer f.Close()
	h := crypto.DefaultHasher()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("error when hashing genesis file: %w", err)
	}
	actualHash := h.Sum(nil)

	// Compare with the flag.
	if !bytes.Equal(genesisHash, actualHash) {
		return fmt.Errorf(
			"--genesis_hash=%X does not match %s hash: %X",
			genesisHash, config.GenesisFile(), actualHash)
	}

	return nil
}

// trapSignal() comes from tmos.TrapSignal
func trapSignal(logger log.Logger, cb func()) {
	var signals = []os.Signal{
		os.Interrupt,
		//syscall.SIGINT,
		//syscall.SIGQUIT,
		//syscall.SIGABRT,
		//syscall.SIGKILL,
		syscall.SIGTERM,
		//syscall.SIGSTOP,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	go func() {
		for sig := range c {
			logger.Info("signal trapped", "msg", log.NewLazySprintf("captured %v, exiting...", sig))
			if cb != nil {
				cb()
			}
			os.Exit(0)
		}
	}()
}
