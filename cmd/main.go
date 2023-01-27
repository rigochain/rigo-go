package main

import (
	"github.com/kysee/arcanus/cmd/commands"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/node"
	"github.com/tendermint/tendermint/libs/cli"
	"path/filepath"
)

func main() {
	commands.RootCmd.AddCommand(
		commands.NewInitFilesCmd(),
		commands.ResetPrivValidatorCmd,
		commands.ResetAllCmd,
		commands.NewRunNodeCmd(node.NewArcanusNode),
		commands.ShowNodeIDCmd,
		commands.NewShowWalletKeyCmd(),
		commands.VersionCmd,
	)

	executor := cli.PrepareBaseCmd(commands.RootCmd, "ARCANUS", filepath.Join(libs.GetHome(), ".arcanus"))
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}
