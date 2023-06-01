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
	validatorWallets []*rigoweb3.Wallet
	wallets          []*rigoweb3.Wallet
	walletsMap       map[rtypes1.AcctKey]*rigoweb3.Wallet
	W0               *rigoweb3.Wallet
	W1               *rigoweb3.Wallet
	amt              = bytes.RandU256IntN(uint256.NewInt(1000))
	baseFee          = uint256.NewInt(1_000_000_000_000_000)
	smallFee         = uint256.NewInt(999_999_999_999_999)
	limitFee         = uint256.NewInt(15_000_000_000_000_000)
	defaultRpcNode   *PeerMock
)

func prepareTest(peers []*PeerMock) {
	for _, peer := range peers {
		// validators
		if w, err := rigoweb3.OpenWallet(libs.NewFileReader(peer.PrivValKeyPath())); err != nil {
			panic(err)
		} else {
			addValidatorWallet(w)
		}

		// wallets
		files, err := os.ReadDir(peer.WalletPath())
		if err != nil {
			panic(err)
		}

		walletsMap = make(map[rtypes1.AcctKey]*rigoweb3.Wallet)

		for _, file := range files {
			if !file.IsDir() {
				if w, err := rigoweb3.OpenWallet(
					libs.NewFileReader(filepath.Join(peer.WalletPath(), file.Name()))); err != nil {
					panic(err)
				} else {
					wallets = append(wallets, w)

					acctKey := rtypes1.ToAcctKey(w.Address())
					walletsMap[acctKey] = w

					//if err := w.SyncAccount(rweb3); err != nil {
					//	panic(err)
					//}
					//fmt.Println(w.Address(), w.GetBalance())
				}
			}
		}
	}

	W0 = wallets[0]
	W1 = wallets[1]
}

func waitTrxResult(txhash []byte, maxTimes int, rweb3 *rigoweb3.RigoWeb3) (*rweb3types.TrxResult, error) {
	for i := 0; i < maxTimes; i++ {
		time.Sleep(100 * time.Millisecond)

		// todo: check why it takes more than 10 secs to fetch a transaction

		txRet, err := rweb3.GetTransaction(txhash)
		if err != nil && strings.Contains(err.Error(), ") not found") {
			continue
		} else if err != nil {
			return nil, err
		} else {
			return txRet, nil
		}
	}
	return nil, xerrors.NewOrdinary("timeout")
}

func waitEvent(query string, cb func(*coretypes.ResultEvent, error) bool) (*sync.WaitGroup, error) {
	subWg := sync.WaitGroup{}
	sub, err := rigoweb3.NewSubscriber(defaultRpcNode.WSEnd)
	if err != nil {
		return nil, err
	}

	subWg.Add(1)
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

func queryValidators(height int, rweb3 *rigoweb3.RigoWeb3) (*coretypes.ResultValidators, error) {
	return rweb3.GetValidators(int64(height), 1, len(validatorWallets))
}

func randWallet() *rigoweb3.Wallet {
	rn := rand.Intn(len(wallets))
	return wallets[rn]
}

func randValidatorWallet() *rigoweb3.Wallet {
	rn := rand.Intn(len(validatorWallets))
	return validatorWallets[rn]
}

func randCommonWallet() *rigoweb3.Wallet {
	for {
		w := randWallet()
		if isValidatorWallet(w) == false {
			return w
		}
	}
}

func saveWallet(w *rigoweb3.Wallet) error {
	path := filepath.Join(defaultRpcNode.WalletPath(), fmt.Sprintf("wk%X.json", w.Address()))
	return w.Save(libs.NewFileWriter(path))
}
