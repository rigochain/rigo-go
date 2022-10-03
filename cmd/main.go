package main

import (
	"github.com/kysee/arcanus/cmd/commands"
	xlib "github.com/kysee/arcanus/libs"
	xnode "github.com/kysee/arcanus/node"
	"github.com/tendermint/tendermint/libs/cli"
	"path/filepath"
)

func main() {
	commands.RootCmd.AddCommand(
		commands.NewInitFilesCmd(xnode.XChainNewNode),
		commands.ResetPrivValidatorCmd,
		commands.ResetAllCmd,
		commands.NewRunNodeCmd(xnode.XChainNewNode),
		commands.ShowNodeIDCmd,
		commands.VersionCmd,
	)

	executor := cli.PrepareBaseCmd(commands.RootCmd, "ARCANUS", filepath.Join(xlib.GetHome(), ".arcanus"))
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}
