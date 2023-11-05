package cmds

import (
	"fmt"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/json"
)

func NewCmd_Account() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "get account",
		RunE:  account,
	}
	return cmd
}

func account(cmd *cobra.Command, args []string) error {
	rweb3 := web3.NewRigoWeb3(web3.NewHttpProvider(rootFlags.RPCUrl))

	acct, err := rweb3.GetAccount(rootFlags.To)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(acct, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
