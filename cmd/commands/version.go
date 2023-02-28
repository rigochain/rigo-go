package commands

import (
	"fmt"
	xver "github.com/rigochain/rigo-go/cmd/version"
	"github.com/spf13/cobra"
)

// VersionCmd ...
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(xver.String())
	},
}
