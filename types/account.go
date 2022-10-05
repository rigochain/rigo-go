package types

import (
	"bytes"
	"encoding/hex"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/crypto"
	"math/big"
	"sort"
	"strings"
)

const AddrSize = 20
const (
	ACCT_COMMON_TYPE int16 = 1 + iota
)

func ToAcctKey(addr Address) AcctKey {
	var key AcctKey
	copy(key[:], addr[:AddrSize])
	return key
}

func ToAddress(key AcctKey) Address {
	addr := make([]byte, AddrSize)
	copy(addr, key[:])
	return addr
}

func Hex2Address(_hex string) (Address, error) {
	if strings.HasPrefix(_hex, "0x") {
		_hex = _hex[2:]
	}
	bzAddr, err := hex.DecodeString(_hex)
	if err != nil {
		return nil, xerrors.Wrap(err)
	}
	if len(bzAddr) != crypto.AddressSize {
		return nil, xerrors.New("error of address length: address length should be 20 bytes")
	}
	return bzAddr, nil
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

type IAccount interface {
	Type() int16
	Key() AcctKey
	GetAddress() Address
	SetName(string)
	GetName() string
	AddNonce()
	GetNonce() uint64
	CheckNonce(uint64) error
	AddBalance(*big.Int) error
	SubBalance(*big.Int) error
	GetBalance() *big.Int
	CheckBalance(*big.Int) error
	SetCode([]byte)
	GetCode() []byte
}

type FindAccountCb func(Address, bool) IAccount

type IAccountFinder interface {
	FindOrNewAccount(Address, bool) IAccount
	FindAccount(Address, bool) IAccount
}

type INonceChecker interface {
	CheckNonce(IAccount, uint64) error
}

type IBalanceChecker interface {
	CheckBalance(IAccount, *big.Int) error
}
