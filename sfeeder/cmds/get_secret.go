package cmds

import (
	"fmt"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/sfeeder/common"
	"github.com/spf13/cobra"
)

var targetAddr []byte

func NewCmd_GetSecret() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"read", "show"},
		Short:   "get secret",
		RunE:    getSecret,
	}
	addGetSecretFlags(cmd)
	return cmd
}

func addGetSecretFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BytesHexVarP(&targetAddr, "addr", "a", targetAddr, "-")
}

func getSecret(cmd *cobra.Command, args []string) error {
	defer func() {
		libs.ClearCredential([]byte(pass))
	}()

	secret, err := common.ReadSecret(targetAddr, []byte(pass), DataDir)
	defer libs.ClearCredential(secret)

	if err != nil {
		return err
	}

	fmt.Println("Secret:", string(secret))
	return nil
}
