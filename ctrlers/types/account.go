package types

import (
	"bytes"
	"encoding/hex"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
	"sort"
	"sync"
)

type Account struct {
	Address types.Address `json:"address"`
	Name    string        `json:"name,omitempty"`
	Nonce   uint64        `json:"nonce,string"`
	Balance *uint256.Int  `json:"balance"`
	Code    []byte        `json:"code,omitempty"`

	mtx sync.RWMutex
}

func NewAccount(addr types.Address) *Account {
	return &Account{
		Address: addr,
		Nonce:   0,
		Balance: uint256.NewInt(0),
	}
}

func NewAccountWithName(addr types.Address, name string) *Account {
	acct := NewAccount(addr)
	acct.Name = name
	return acct
}

func (acct *Account) GetAddress() types.Address {
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

func (acct *Account) CheckNonce(n uint64) xerrors.XError {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	if acct.Nonce+1 != n {
		return xerrors.ErrInvalidNonce
	}
	return nil
}

func (acct *Account) AddBalance(amt *uint256.Int) xerrors.XError {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}
	_ = acct.Balance.Add(acct.Balance, amt)

	return nil
}

func (acct *Account) SubBalance(amt *uint256.Int) xerrors.XError {
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

func (acct *Account) GetBalance() *uint256.Int {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	return new(uint256.Int).Set(acct.Balance)
}

func (acct *Account) CheckBalance(amt *uint256.Int) xerrors.XError {
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

func (acct *Account) Type() int16 {
	return types.ACCT_COMMON_TYPE
}

func (acct *Account) Key() ledger.LedgerKey {
	acct.mtx.RLock()
	acct.mtx.RUnlock()

	return acct.Address.Array32()
}

func (acct *Account) Encode() ([]byte, xerrors.XError) {
	if bz, err := proto.Marshal(&AcctProto{
		Address:  acct.Address,
		Name:     acct.Name,
		Nonce:    acct.Nonce,
		XBalance: acct.Balance.Bytes(),
		XCode:    acct.Code,
	}); err != nil {
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (acct *Account) Decode(d []byte) xerrors.XError {
	pm := &AcctProto{}
	if err := proto.Unmarshal(d, pm); err != nil {
		return xerrors.From(err)
	}

	acct.Address = pm.Address
	acct.Name = pm.Name
	acct.Nonce = pm.Nonce
	acct.Balance = new(uint256.Int).SetBytes(pm.XBalance)
	acct.Code = pm.XCode
	return nil
}

var _ ledger.ILedgerItem = (*Account)(nil)

////

type AcctKey [types.AddrSize]byte

func RandAddrKey() AcctKey {
	var k AcctKey
	copy(k[:], abytes.RandBytes(types.AddrSize))
	return k
}

func ToAcctKey(addr types.Address) AcctKey {
	var key AcctKey
	copy(key[:], addr[:types.AddrSize])
	return key
}

// MarshalText() is needed to use AcctKey as key of map

func (ak AcctKey) MarshalText() ([]byte, error) {
	s := hex.EncodeToString(ak[:])
	return []byte(s), nil
}

func (ak AcctKey) Address() types.Address {
	addr := make([]byte, types.AddrSize)
	copy(addr, ak[:])
	return addr
}

func (ak AcctKey) String() string {
	return hex.EncodeToString(ak[:])
}

type AcctKeyList []AcctKey

func (a AcctKeyList) Len() int {
	return len(a)
}

func (a AcctKeyList) Less(i, j int) bool {
	ret := bytes.Compare(a[i][:], a[j][:])
	return ret > 0
}

func (a AcctKeyList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

var _ sort.Interface = AcctKeyList(nil)
