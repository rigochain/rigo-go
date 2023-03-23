package test

import (
	"fmt"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs"
	rigoweb3 "github.com/rigochain/rigo-go/libs/web3"
	rweb3types "github.com/rigochain/rigo-go/libs/web3/types"
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
	rweb3           = rigoweb3.NewRigoWeb3(rigoweb3.NewHttpProvider("http://localhost:26657"))
	home, _         = os.UserHomeDir()
	VALWALLETPATH   = filepath.Join(home, "rigo_localnet_0/config/priv_validator_key.json")
	WALKEYDIR       = filepath.Join(home, "rigo_localnet_0/walkeys")
	TESTPASS        = []byte("1")
	validatorWallet *rigoweb3.Wallet
	wallets         []*rigoweb3.Wallet
	walletsMap      map[types.AcctKey]*rigoweb3.Wallet
	W0              *rigoweb3.Wallet
	W1              *rigoweb3.Wallet
	amt             = bytes.RandBigIntN(big.NewInt(1000))
	gas             = big.NewInt(10)
)

func init() {

	files, err := os.ReadDir(WALKEYDIR)
	if err != nil {
		panic(err)
	}

	walletsMap = make(map[types.AcctKey]*rigoweb3.Wallet)

	if w, err := rigoweb3.OpenWallet(libs.NewFileReader(VALWALLETPATH)); err != nil {
		panic(err)
	} else {
		validatorWallet = w
	}

	for _, file := range files {
		if !file.IsDir() {
			if w, err := rigoweb3.OpenWallet(
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

func waitTrxResult(txhash []byte, maxTimes int) (*rweb3types.TrxResult, error) {
	for i := 0; i < maxTimes; i++ {
		time.Sleep(100 * time.Millisecond)

		// todo: check why it takes more than 10 secs to fetch a transaction

		txRet, err := rweb3.GetTransaction(txhash)
		if err != nil && !strings.Contains(err.Error(), ") not found") {
			return nil, err
		} else if err == nil {
			return txRet, nil
		}
	}
	return nil, xerrors.New("timeout")
}

func randWallet() *rigoweb3.Wallet {
	rn := rand.Intn(len(wallets))
	return wallets[rn]
}

func randCommonWallet() *rigoweb3.Wallet {
	for {
		w := randWallet()
		if bytes.Compare(w.Address(), validatorWallet.Address()) != 0 {
			return w
		}
	}
}

func saveRandWallet(w *rigoweb3.Wallet) error {
	path := filepath.Join(WALKEYDIR, fmt.Sprintf("wk%X.json", w.Address()))
	return w.Save(libs.NewFileWriter(path))
}
