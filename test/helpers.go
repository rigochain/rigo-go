package test

import (
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/libs/client"
	"github.com/rigochain/rigo-go/libs/client/rpc"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/tendermint/tendermint/libs/rand"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	home, _         = os.UserHomeDir()
	VALWALLETPATH   = filepath.Join(home, ".rigo/config/priv_validator_key.json")
	WALKEYDIR       = filepath.Join(home, ".rigo/walkeys")
	TESTPASS        = []byte("1111")
	validatorWallet *client.Wallet
	wallets         []*client.Wallet
	walletsMap      map[types.AcctKey]*client.Wallet
	W0              *client.Wallet
	W1              *client.Wallet
	amt             = bytes.RandBigIntN(big.NewInt(1000))
	gas             = big.NewInt(10)
)

func init() {

	files, err := os.ReadDir(WALKEYDIR)
	if err != nil {
		panic(err)
	}

	walletsMap = make(map[types.AcctKey]*client.Wallet)

	if w, err := client.OpenWallet(libs.NewFileReader(VALWALLETPATH)); err != nil {
		panic(err)
	} else {
		validatorWallet = w
	}

	for _, file := range files {
		if !file.IsDir() {
			if w, err := client.OpenWallet(
				libs.NewFileReader(filepath.Join(WALKEYDIR, file.Name()))); err != nil {
				panic(err)
			} else {
				wallets = append(wallets, w)

				acctKey := types.ToAcctKey(w.Address())
				walletsMap[acctKey] = w
			}
		}
	}
	W0 = wallets[0]
	W1 = wallets[1]
}

func waitTrxResult(txhash []byte, maxTimes int) (*rpc.TrxResult, error) {
	for i := 0; i < maxTimes; i++ {
		time.Sleep(100 * time.Millisecond)

		// todo: check why it takes more than 10 secs to fetch a transaction

		txRet, err := rpc.GetTransaction(txhash)
		if err != nil && !strings.Contains(err.Error(), ") not found") {
			return nil, err
		} else if err == nil {
			return txRet, nil
		}
	}
	return nil, xerrors.New("timeout")
}

func randWallet() *client.Wallet {
	rn := rand.Intn(len(wallets))
	return wallets[rn]
}

func randCommonWallet() *client.Wallet {
	for {
		w := randWallet()
		if bytes.Compare(w.Address(), validatorWallet.Address()) != 0 {
			return w
		}
	}
}
