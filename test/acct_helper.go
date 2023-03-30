package test

import (
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"math/big"
	"sync"
)

type acctHelper struct {
	w *web3.Wallet

	originBalance *big.Int
	originNonce   uint64

	txHashes        map[string]types.Address
	spentGas        *big.Int
	expectedBalance *big.Int
	expectedNonce   uint64

	mtx sync.RWMutex
}

func newAcctHelper(w *web3.Wallet) *acctHelper {
	return &acctHelper{
		w:               w,
		originBalance:   w.GetBalance(),
		originNonce:     w.GetNonce(),
		txHashes:        make(map[string]types.Address),
		spentGas:        big.NewInt(0),
		expectedBalance: w.GetBalance(),
		expectedNonce:   w.GetNonce(),
	}
}

func (obj *acctHelper) addTxHashOfAddr(txhash string, addr types.Address) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.txHashes[txhash] = addr
}

func (obj *acctHelper) addSpentGas(d *big.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.spentGas = new(big.Int).Add(obj.spentGas, d)
}

func (obj *acctHelper) addExpectedBalance(d *big.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.expectedBalance = new(big.Int).Add(obj.expectedBalance, d)
}

func (obj *acctHelper) subExpectedBalance(d *big.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.expectedBalance = new(big.Int).Sub(obj.expectedBalance, d)
}

func (obj *acctHelper) addExpectedNonce() {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.expectedNonce++
}

var senderAcctHelpers = make(map[string]*acctHelper)
var gmtx = &sync.Mutex{}

func addSenderAccotHelper(k string, v *acctHelper) {
	gmtx.Lock()
	defer gmtx.Unlock()

	senderAcctHelpers[k] = v
}
