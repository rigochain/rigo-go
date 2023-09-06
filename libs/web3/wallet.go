package web3

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"io"
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

func (w *Wallet) Save(wr io.Writer) error {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	_, err := w.wkey.Save(wr)
	return err
}

func (w *Wallet) Clone() *Wallet {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return &Wallet{
		wkey: w.wkey,
		acct: w.acct.Clone(),
	}
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

func (w *Wallet) AddNonce() {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.addNonce()
}

func (w *Wallet) addNonce() {
	w.acct.AddNonce()
}

func (w *Wallet) GetNonce() uint64 {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.acct.GetNonce()
}

func (w *Wallet) GetBalance() *uint256.Int {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	if w.acct == nil {
		return uint256.NewInt(0)
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

func (w *Wallet) SignTrxRLP(tx *types2.Trx) (bytes.HexBytes, bytes.HexBytes, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	if txbz, err := rlp.EncodeToBytes(tx); err != nil {
		return nil, nil, err
	} else if sig, err := w.wkey.Sign(txbz); err != nil {
		return nil, nil, err
	} else {
		tx.Sig = sig
		return sig, txbz, nil
	}
}

func (w *Wallet) TransferAsync(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxTransfer(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxAsync(tx, rweb3)
}

func (w *Wallet) TransferSync(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxTransfer(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxSync(tx, rweb3)
}

func (w *Wallet) TransferCommit(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	tx := NewTrxTransfer(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxCommit(tx, rweb3)
}

func (w *Wallet) StakingAsync(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxStaking(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxAsync(tx, rweb3)
}

func (w *Wallet) StakingSync(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxStaking(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxSync(tx, rweb3)
}

func (w *Wallet) StakingCommit(to types.Address, gas, amt *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	tx := NewTrxStaking(
		w.Address(), to,
		w.acct.GetNonce(),
		gas, amt,
	)
	return w.SendTxCommit(tx, rweb3)
}

func (w *Wallet) WithdrawAync(fee, req *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxWithdraw(w.Address(), w.Address(), w.acct.GetNonce(), fee, req)
	return w.SendTxAsync(tx, rweb3)
}

func (w *Wallet) WithdrawSync(fee, req *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxWithdraw(w.Address(), w.Address(), w.acct.GetNonce(), fee, req)
	return w.SendTxSync(tx, rweb3)
}

func (w *Wallet) WithdrawCommit(fee, req *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	tx := NewTrxWithdraw(w.Address(), w.Address(), w.acct.GetNonce(), fee, req)
	return w.SendTxCommit(tx, rweb3)
}

func (w *Wallet) SendTxAsync(tx *types2.Trx, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionAsync(tx)
	}
}

func (w *Wallet) SendTxSync(tx *types2.Trx, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionSync(tx)
	}
}

func (w *Wallet) SendTxCommit(tx *types2.Trx, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionCommit(tx)
	}
}

func (w *Wallet) SetDocSync(name, url string, gas *uint256.Int, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxSetDoc(w.Address(), w.acct.GetNonce(), gas, name, url)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionSync(tx)
	}
}

func (w *Wallet) syncAccount(rweb3 *RigoWeb3) error {
	if acct, err := rweb3.GetAccount(w.wkey.Address); err != nil {
		return err
	} else {
		w.acct = acct
	}
	return nil
}

func (w *Wallet) SyncAccount(rweb3 *RigoWeb3) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncAccount(rweb3)
}

func (w *Wallet) syncNonce(rweb3 *RigoWeb3) error {
	return w.syncAccount(rweb3)
}

func (w *Wallet) SyncNonce(rweb3 *RigoWeb3) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncNonce(rweb3)
}

func (w *Wallet) syncBalance(rweb3 *RigoWeb3) error {
	return w.syncAccount(rweb3)
}

func (w *Wallet) SyncBalance(rweb3 *RigoWeb3) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.syncBalance(rweb3)
}

func (w *Wallet) ProposalSync(gas *uint256.Int, msg string, start, period int64, optType int32, options []byte, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxProposal(
		w.Address(),
		types.ZeroAddress(),
		w.acct.GetNonce(),
		gas, msg, start, period, optType, options,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionSync(tx)
	}
}

func (w *Wallet) ProposalSyncCommit(gas *uint256.Int, msg string, start, period int64, optType int32, options []byte, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	tx := NewTrxProposal(
		w.Address(),
		types.ZeroAddress(),
		w.acct.GetNonce(),
		gas, msg, start, period, optType, options,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionCommit(tx)
	}
}

func (w *Wallet) VotingSync(gas *uint256.Int, txHash bytes.HexBytes, choice int32, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTx, error) {
	tx := NewTrxVoting(
		w.Address(),
		types.ZeroAddress(),
		w.acct.GetNonce(),
		gas, txHash, choice,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionSync(tx)
	}
}

func (w *Wallet) VotingSyncCommit(gas *uint256.Int, txHash bytes.HexBytes, choice int32, rweb3 *RigoWeb3) (*coretypes.ResultBroadcastTxCommit, error) {
	tx := NewTrxVoting(
		w.Address(),
		types.ZeroAddress(),
		w.acct.GetNonce(),
		gas, txHash, choice,
	)
	if _, _, err := w.SignTrx(tx); err != nil {
		return nil, err
	} else {
		return rweb3.SendTransactionCommit(tx)
	}
}
