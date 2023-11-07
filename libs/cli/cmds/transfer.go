package cmds

import (
	"errors"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/spf13/cobra"
)

func NewCmd_Transfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer RIGO coin",
		RunE:  transfer,
	}
	return cmd
}

func transfer(cmd *cobra.Command, args []string) error {
	rweb3 := web3.NewRigoWeb3(web3.NewHttpProvider(rootFlags.RPCUrl))

	// get sender's wallet key
	if rootFlags.From == "" {
		return errors.New("please set the wallet key file of sender")
	}
	w, err := web3.OpenWallet(libs.NewFileReader(rootFlags.From))
	if err != nil {
		return err
	}

	w.SyncAccount(rweb3)
	w.Unlock([]byte("1111"))

	ret, err := w.TransferCommit(rootFlags.To, rootFlags.Gas, uint256.MustFromDecimal(rootFlags.GasPrice), uint256.MustFromDecimal(rootFlags.Amount), rweb3)
	if err != nil {
		return err
	}
	if ret.CheckTx.Code != 0 {
		return errors.New(ret.CheckTx.Log)
	}
	if ret.DeliverTx.Code != 0 {
		return errors.New(ret.DeliverTx.Log)
	}

	return nil
}
