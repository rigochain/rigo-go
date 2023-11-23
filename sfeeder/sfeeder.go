package main

import (
	"github.com/rigochain/rigo-go/sfeeder/cmds"
	"os"
)

func main() {
	baseCmd := cmds.NewCmd_Base()
	if err := baseCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
