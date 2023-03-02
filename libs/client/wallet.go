package client

import (
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/client/rpc"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"io"
	"math/big"
	"sync"
)

type Wallet struct {
	wkey *crypto.WalletKey
	acct *types2.Account

	mtx sync.RWMutex
}

func NewWallet(s []byte) *Wallet {
	prvKey := tmsecp256k1.GenPrivKey()
	wkey := crypto.NewWalletKey(prvKey, s)
	return &Wallet{
		wkey: wkey,
		acct: types2.NewAccount(wkey.Address),
	}
}

func OpenWallet(r io.Reader) (*Wallet, error) {
	wk, err := crypto.OpenWalletKey(r)
	if err != nil {
		return nil, err
	}

	ret := &Wallet{
		wkey: wk,
	}

	//if err := ret.syncAccount(); err != nil {
	//	return nil, err
	//}
	return ret, nil
}

func (w *Wallet) Address() types.Address {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.wkey.Address
}

func (w *Wallet) GetAccount() *types2.Account {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.acct
}

func (w *Wallet) syncAccount() error {
	if acct, err := rpc.GetAccount(w.wkey.Address); err != nil {
		return err
	} else {
		w.acct = acct
	}
	return nil
}

func (w *Wallet) SyncAccount() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncAccount()
}

func (w *Wallet) syncNonce() error {
	return w.syncAccount()
}

func (w *Wallet) AddNonce() {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.addNonce()
}

func (w *Wallet) addNonce() {
	w.acct.AddNonce()
}

func (w *Wallet) SyncNonce() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncNonce()
}

func (w *Wallet) GetNonce() uint64 {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.acct.GetNonce()
}

func (w *Wallet) syncBalance() error {
	return w.syncAccount()
}

func (w *Wallet) SyncBalance() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncBalance()
}

func (w *Wallet) GetBalance() *big.Int {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	if w.acct == nil {
		return big.NewInt(0)
	} else {
		return w.acct.GetBalance()
	}
}

func (w *Wallet) Lock() {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.wkey.Lock()
}

func (w *Wallet) GetPubKey() bytes.HexBytes {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.wkey.PubKey()
}

func (w *Wallet) Unlock(s []byte) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.wkey.Unlock(s)
}

func (w *Wallet) SignTrx(tx *types2.Trx) (bytes.HexBytes, bytes.HexBytes, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	if txbz, err := tx.Encode(); err != nil {
		return nil, nil, err
	} else if sig, err := w.wkey.Sign(txbz); err != nil {
		return nil, nil, err
	} else {
		tx.Sig = sig
		return sig, txbz, nil
	}
}

func (w *Wallet) TransferSync(to types.Address, gas, amt *big.Int) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxTransfer(
		w.Address(), to,
		w.acct.GetNonce()+1,
		gas, amt,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rpc.SendTransactionSync(tx)
	}
}

func (w *Wallet) StakingSync(to types.Address, gas, amt *big.Int) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxStaking(
		w.Address(), to,
		w.acct.GetNonce()+1,
		gas, amt,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rpc.SendTransactionSync(tx)
	}
}
