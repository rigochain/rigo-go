package commands

import (
	"fmt"
	"github.com/kysee/arcanus/genesis"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/crypto"
	xnode "github.com/kysee/arcanus/node"
	"github.com/spf13/cobra"
	cfg "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
	"path/filepath"
)

//// InitFilesCmd initialises a fresh Tendermint Core instance.
//var InitFilesCmd = &cobra.Command{
//	Use:   "init",
//	Short: "Initialize a arcanus",
//	RunE:  initFiles,
//}

var (
	chainID       = "mainnet"
	walkeyCnt     = 9
	privValSecret string
)

// NewRunNodeCmd returns the command that allows the CLI to start a node.
// It can be used with a custom PrivValidator and in-process ABCI application.
func NewInitFilesCmd(nodeProvider xnode.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a arcanus",
		RunE:  initFiles,
	}
	AddInitFlags(cmd)
	return cmd
}

func AddInitFlags(cmd *cobra.Command) {
	// bind flags
	cmd.Flags().StringVar(
		&chainID,
		"chain_id",
		chainID, // default name
		"the id of chain to generate (e.g. mainnet, testnet, devnet and others)")
	cmd.Flags().IntVar(
		&walkeyCnt,
		"holders",
		walkeyCnt, // default value is 9
		"the number of holder's accounts to generate. "+
			"if you create a new genesis of your own blockchain, "+
			"you need to generate accounts of genesis holders and "+
			"these accounts will be saved at $XCNCHOME/walkeys",
	)
	cmd.Flags().StringVar(
		&privValSecret,
		"priv_validator_secret",
		"",
		"passphrase to encrypt and decrypt a private key in priv_validator_key.json",
	)
}

func initFiles(cmd *cobra.Command, args []string) error {
	return initFilesWithConfig(config)
}

func initFilesWithConfig(config *cfg.Config) error {
	var s []byte
	if privValSecret != "" {
		s = []byte(privValSecret)
		privValSecret = ""
	} else {
		s = libs.ReadCredential(fmt.Sprintf("Passphrase for %v: ", filepath.Base(config.PrivValidatorKeyFile())))
	}
	defer libs.ClearCredential(s)

	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *crypto.SFilePV
	if tmos.FileExists(privValKeyFile) {
		pv = crypto.LoadSFilePV(privValKeyFile, privValStateFile, s)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
		//pv.SaveWith(s) // encrypt with new driven key.
	} else {
		pv = crypto.GenSFilePV(privValKeyFile, privValStateFile)
		pv.SaveWith(s)
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if tmos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		var err error
		var genDoc *tmtypes.GenesisDoc
		if chainID == "mainnet" {
			if genDoc, err = genesis.MainnetGenesisDoc(); err != nil {
				return err
			}
		} else if chainID == "testnet" {
			if genDoc, err = genesis.TestnetGenesisDoc(); err != nil {
				return err
			}
		} else if chainID == "devnet" {
			if genDoc, err = genesis.DevnetGenesisDoc(); err != nil {
				return err
			}
		} else { // anything (e.g. loclanet)
			defaultWalkeyDirPath := filepath.Join(config.RootDir, crypto.DefaultWalletKeyDir)
			err := tmos.EnsureDir(defaultWalkeyDirPath, crypto.DefaultWalletKeyDirPerm)
			if err != nil {
				return err
			}

			walkeys, err := crypto.CreateWalletKeyFiles(s, walkeyCnt, defaultWalkeyDirPath)
			if err != nil {
				return err
			}

			pvWalKey, err := crypto.OpenWalletKey(libs.NewFileReader(privValKeyFile))
			if err != nil {
				return err
			}
			_, err = pvWalKey.Save(
				libs.NewFileWriter(
					filepath.Join(defaultWalkeyDirPath, fmt.Sprintf("wk%X.json", pvWalKey.Address))))
			if err != nil {
				return err
			}

			pubKey, err := pv.GetPubKey()
			if err != nil {
				return fmt.Errorf("can't get pubkey: %w", err)
			}
			valset := []tmtypes.GenesisValidator{{
				Address: pubKey.Address(),
				PubKey:  pubKey,
				Power:   10,
			}}

			walkeys = append(walkeys, pvWalKey)
			holders := make([]*genesis.GenesisAssetHolder, len(walkeys))
			for i, wk := range walkeys {
				if err := wk.Unlock(s); err != nil {
					return err
				}
				holders[i] = &genesis.GenesisAssetHolder{
					Address: wk.Address,
					Balance: "100000000000000000000000000", // 100_000_000_000000000000000000
				}
			}
			defer func() {
				for _, wk := range walkeys {
					wk.Lock()
				}
			}()

			govRule := genesis.DefaultGenesisGovRule()
			genDoc, err = genesis.NewGenesisDoc(chainID, valset, holders, govRule)
			if err != nil {
				return err
			}

		}
		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}
