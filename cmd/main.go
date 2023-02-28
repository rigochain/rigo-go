package main

import (
	"github.com/rigochain/rigo-go/cmd/commands"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/node"
	"github.com/tendermint/tendermint/libs/cli"
	"path/filepath"
)

func main() {
	commands.RootCmd.AddCommand(
		commands.NewInitFilesCmd(),
		commands.ResetPrivValidatorCmd,
		commands.ResetAllCmd,
		commands.NewRunNodeCmd(node.NewRigoNode),
		commands.ShowNodeIDCmd,
		commands.NewShowWalletKeyCmd(),
		commands.VersionCmd,
	)

	executor := cli.PrepareBaseCmd(commands.RootCmd, "RIGO", filepath.Join(libs.GetHome(), ".rigo"))
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}
