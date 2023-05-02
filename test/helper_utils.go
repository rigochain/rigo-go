package test

import (
	"fmt"
	"github.com/holiman/uint256"
	rtypes1 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs"
	rigoweb3 "github.com/rigochain/rigo-go/libs/web3"
	rweb3types "github.com/rigochain/rigo-go/libs/web3/types"
	rtypes0 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/rand"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	rweb3            *rigoweb3.RigoWeb3
	validatorWallets []*rigoweb3.Wallet
	wallets          []*rigoweb3.Wallet
	walletsMap       map[rtypes1.AcctKey]*rigoweb3.Wallet
	W0               *rigoweb3.Wallet
	W1               *rigoweb3.Wallet
	amt              = bytes.RandU256IntN(uint256.NewInt(1000))
	gas              = uint256.NewInt(10)
)

func prepareTest() {
	rweb3 = rigoweb3.NewRigoWeb3(rigoweb3.NewHttpProvider(rpcURL))

	files, err := os.ReadDir(walletPath())
	if err != nil {
		panic(err)
	}

	walletsMap = make(map[rtypes1.AcctKey]*rigoweb3.Wallet)

	if w, err := rigoweb3.OpenWallet(libs.NewFileReader(privValKeyPath())); err != nil {
		panic(err)
	} else {
		addValidatorWallet(w)
	}

	for _, file := range files {
		if !file.IsDir() {
			if w, err := rigoweb3.OpenWallet(
				libs.NewFileReader(filepath.Join(walletPath(), file.Name()))); err != nil {
				panic(err)
			} else {
				wallets = append(wallets, w)

				acctKey := rtypes1.ToAcctKey(w.Address())
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
	return nil, xerrors.NewOrdinary("timeout")
}

func randWallet() *rigoweb3.Wallet {
	rn := rand.Intn(len(wallets))
	return wallets[rn]
}

func addValidatorWallet(w *rigoweb3.Wallet) {
	gmtx.Lock()
	defer gmtx.Unlock()

	validatorWallets = append(validatorWallets, w)
}

func isValidatorWallet(w *rigoweb3.Wallet) bool {
	return isValidator(w.Address())
}

func isValidator(addr rtypes0.Address) bool {
	for _, vw := range validatorWallets {
		if bytes.Compare(vw.Address(), addr) == 0 {
			return true
		}
	}
	return false
}

func validators(height int64) (*coretypes.ResultValidators, error) {
	return rweb3.GetValidators(height, 1, len(validatorWallets))
}

func randCommonWallet() *rigoweb3.Wallet {
	for {
		w := randWallet()
		if isValidatorWallet(w) == false {
			return w
		}
	}
}

func saveRandWallet(w *rigoweb3.Wallet) error {
	path := filepath.Join(walletPath(), fmt.Sprintf("wk%X.json", w.Address()))
	return w.Save(libs.NewFileWriter(path))
}

func waitEvent(query string, cb func(*coretypes.ResultEvent, error) bool) (*sync.WaitGroup, error) {
	subWg := sync.WaitGroup{}
	sub, err := rigoweb3.NewSubscriber(wsEndpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		sub.Stop()
	}()

	err = sub.Start(query, func(sub *rigoweb3.Subscriber, result []byte) {

		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		if cb(event, err) {
			sub.Stop()
			subWg.Done()
		}
	})
	if err != nil {
		return nil, err
	}

	return &subWg, nil
}
