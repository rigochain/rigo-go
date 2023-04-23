package test

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"sync"
)

type acctHelper struct {
	w *web3.Wallet

	originBalance *uint256.Int
	originNonce   uint64

	txHashes        map[string]types.Address
	spentGas        *uint256.Int
	expectedBalance *uint256.Int
	expectedNonce   uint64

	mtx sync.RWMutex
}

func newAcctHelper(w *web3.Wallet) *acctHelper {
	return &acctHelper{
		w:               w,
		originBalance:   w.GetBalance(),
		originNonce:     w.GetNonce(),
		txHashes:        make(map[string]types.Address),
		spentGas:        uint256.NewInt(0),
		expectedBalance: w.GetBalance(),
		expectedNonce:   w.GetNonce(),
	}
}

func (obj *acctHelper) addTxHashOfAddr(txhash string, addr types.Address) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.txHashes[txhash] = addr
}

func (obj *acctHelper) addSpentGas(d *uint256.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	_ = obj.spentGas.Add(obj.spentGas, d)
}

func (obj *acctHelper) addExpectedBalance(d *uint256.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	_ = obj.expectedBalance.Add(obj.expectedBalance, d)
}

func (obj *acctHelper) subExpectedBalance(d *uint256.Int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	_ = obj.expectedBalance.Sub(obj.expectedBalance, d)
}

func (obj *acctHelper) addExpectedNonce() {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.expectedNonce++
}

var senderAcctHelpers = make(map[string]*acctHelper)
var gmtx = &sync.Mutex{}

func addSenderAcctHelper(k string, v *acctHelper) {
	gmtx.Lock()
	defer gmtx.Unlock()

	senderAcctHelpers[k] = v
}
