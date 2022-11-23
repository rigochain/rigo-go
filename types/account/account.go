package account

import (
	"errors"
	"fmt"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
	"sync"
)

type Account struct {
	Address Address  `json:"address"`
	Name    string   `json:"name,omitempty"`
	Nonce   uint64   `json:"nonce"`
	Balance *big.Int `json:"balance,string"`
	Code    []byte   `json:"code,omitempty"`

	mtx sync.RWMutex
}

func NewAccount(addr Address) *Account {
	return &Account{
		Address: addr,
		Nonce:   0,
		Balance: big.NewInt(0),
	}
}

func NewAccountWithName(addr Address, name string) *Account {
	return &Account{
		Address: addr,
		Name:    name,
		Nonce:   0,
		Balance: big.NewInt(0),
	}
}

func (acct *Account) Type() int16 {
	return ACCT_COMMON_TYPE
}

func (acct *Account) Key() AcctKey {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return ToAcctKey(acct.GetAddress())
}

func (acct *Account) GetAddress() Address {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return acct.Address
}

func (acct *Account) SetName(s string) {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	acct.Name = s
}

func (acct *Account) GetName() string {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return acct.Name
}

func (acct *Account) AddNonce() {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	acct.Nonce++
}

func (acct *Account) GetNonce() uint64 {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return acct.Nonce
}

func (acct *Account) CheckNonce(n uint64) error {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	if acct.Nonce+1 != n {
		return xerrors.ErrInvalidNonce.Wrap(errors.New(fmt.Sprintf("address: %v, expected: %v, actual:%v", acct.Address, acct.Nonce+1, n)))
	}
	return nil
}

func (acct *Account) AddBalance(amt *big.Int) error {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}
	_ = acct.Balance.Add(acct.Balance, amt)

	return nil
}

func (acct *Account) SubBalance(amt *big.Int) error {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}
	if amt.Cmp(acct.Balance) > 0 {
		return xerrors.ErrInsufficientFund
	}

	_ = acct.Balance.Sub(acct.Balance, amt)
	return nil
}

func (acct *Account) GetBalance() *big.Int {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return new(big.Int).Set(acct.Balance)
}

func (acct *Account) CheckBalance(amt *big.Int) error {
	if amt.Cmp(acct.Balance) > 0 {
		return xerrors.ErrInsufficientFund
	}
	return nil
}

func (acct *Account) SetCode(c []byte) {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	acct.Code = c
}

func (acct *Account) GetCode() []byte {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return acct.Code
}

type FindAccountCb func(Address, bool) *Account

type IAccountFinder interface {
	FindOrNewAccount(Address, bool) *Account
	FindAccount(Address, bool) *Account
}

type INonceChecker interface {
	CheckNonce(*Account, uint64) error
}

type IBalanceChecker interface {
	CheckBalance(*Account, *big.Int) error
}
