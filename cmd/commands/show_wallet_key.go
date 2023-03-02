package commands

import (
	"fmt"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/json"
	"os"
	"path/filepath"
	"strings"
)

var (
	wkSecret string
)

func AddShowWalletKeyCmdFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&wkSecret,
		"secret",
		"",
		"passphrase to encrypt and decrypt a private key in a wallet key files",
	)
}

func NewShowWalletKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-wallet-key",
		Aliases: []string{"show_wallet_key"},
		Short:   "Show wallet key file",
		RunE:    showWalletKey,
		PreRun:  deprecateSnakeCase,
	}

	AddShowWalletKeyCmdFlag(cmd)

	return cmd
}

func showWalletKey(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "~") {
			if home, err := os.UserHomeDir(); err != nil {
				return err
			} else {
				arg = strings.Replace(arg, "~", home, 1)
			}

		}
		fileInfo, err := os.Stat(arg)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			if err := showWalletKeyDir(arg); err != nil {
				return err
			}
		} else {
			if err := showWalletKeyFile(arg); err != nil {
				return err
			}
		}
	}
	return nil
}

func showWalletKeyDir(path string) error {
	err := filepath.WalkDir(path, func(entry string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Println("it is directory", entry)
		} else if err := showWalletKeyFile(entry); err != nil {
			return err
		}
		fmt.Println("---\n")
		return nil
	})
	return err
}

func showWalletKeyFile(path string) error {

	if wk, err := crypto.OpenWalletKey(libs.NewFileReader(path)); err != nil {
		return err
	} else {
		var s []byte
		if wkSecret != "" {
			s = []byte(wkSecret)
		} else {
			s = libs.ReadCredential(fmt.Sprintf("Passphrase for %v: ", filepath.Base(path)))
		}

		if err := wk.Unlock(s); err != nil {
			return err
		}
		defer wk.Lock()

		tmp := &struct {
			*crypto.WalletKey `json:"walletKey"`
			PrvKey            bytes.HexBytes `json."prvKey"`
			PubKey            bytes.HexBytes `json."pubKey"`
		}{
			WalletKey: wk,
			PrvKey:    wk.PrvKey(),
			PubKey:    wk.PubKey(),
		}
		if bz, err := json.MarshalIndent(tmp, "", " "); err != nil {
			return err
		} else {
			fmt.Println(string(bz))
		}
	}
	return nil

}
