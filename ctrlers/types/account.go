package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
	"math/big"
	"sort"
	"sync"
)

type Account struct {
	Address types.Address
	Name    string
	Nonce   uint64
	Balance *big.Int
	Code    []byte

	mtx sync.RWMutex
}

func NewAccount(addr types.Address) *Account {
	return &Account{
		Address: addr,
		Nonce:   0,
		Balance: big.NewInt(0),
	}
}

func NewAccountWithName(addr types.Address, name string) *Account {
	return &Account{
		Address: addr,
		Name:    name,
		Nonce:   0,
		Balance: big.NewInt(0),
	}
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
		return xerrors.ErrInvalidNonce.Wrap(errors.New(fmt.Sprintf("address: %v, expected: %v, actual:%v", acct.Address, acct.Nonce+1, n)))
	}
	return nil
}

func (acct *Account) AddBalance(amt *big.Int) xerrors.XError {
	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}
	_ = acct.Balance.Add(acct.Balance, amt)

	return nil
}

func (acct *Account) SubBalance(amt *big.Int) xerrors.XError {
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

func (acct *Account) CheckBalance(amt *big.Int) xerrors.XError {
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
		return nil, xerrors.NewFrom(err)
	} else {
		return bz, nil
	}
}

func (acct *Account) Decode(d []byte) xerrors.XError {
	pm := &AcctProto{}
	if err := proto.Unmarshal(d, pm); err != nil {
		return xerrors.NewFrom(err)
	}

	acct.Address = pm.Address
	acct.Name = pm.Name
	acct.Nonce = pm.Nonce
	acct.Balance = new(big.Int).SetBytes(pm.XBalance)
	acct.Code = pm.XCode
	return nil
}

//type Account struct {
//	Address types.Address `json:"address"`
//	Name    string        `json:"name,omitempty"`
//	Nonce   uint64        `json:"nonce"`
//	Balance *big.Int      `json:"balance"`
//	Code    []byte        `json:"code,omitempty"`
//
//	mtx sync.RWMutex
//}

func (acct *Account) MarshalJSON() ([]byte, error) {
	acct.mtx.RLock()
	defer acct.mtx.RUnlock()

	_acct := &struct {
		Address types.Address `json:"address"`
		Name    string        `json:"name,omitempty"`
		Nonce   uint64        `json:"nonce"`
		Balance string        `json:"balance"`
		Code    []byte        `json:"code,omitempty"`
	}{
		Address: acct.Address,
		Name:    acct.Name,
		Nonce:   acct.Nonce,
		Balance: acct.Balance.String(),
		Code:    acct.Code,
	}

	return json.Marshal(_acct)
}

func (acct *Account) UnmarshalJSON(d []byte) error {
	_acct := &struct {
		Address types.Address `json:"address"`
		Name    string        `json:"name,omitempty"`
		Nonce   uint64        `json:"nonce"`
		Balance string        `json:"balance"`
		Code    []byte        `json:"code,omitempty"`
	}{}
	if err := json.Unmarshal(d, _acct); err != nil {
		return err
	}

	acct.mtx.Lock()
	defer acct.mtx.Unlock()

	bal, ok := new(big.Int).SetString(_acct.Balance, 10)
	if !ok {
		return errors.New("error in converting string to big.Int")
	}

	acct.Address = _acct.Address
	acct.Name = _acct.Name
	acct.Nonce = _acct.Nonce
	acct.Balance = bal
	acct.Code = _acct.Code
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
