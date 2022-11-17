package account

import (
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types/account"
	"math/big"
)

func EncodeAccount(acct *account.Account) ([]byte, error) {
	return proto.Marshal(ToProto(acct))
}

func DecodeAccount(bz []byte) (*account.Account, error) {
	pm := &account.AcctProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return nil, err
	}
	return FromProto(pm), nil
}

func ToProto(acct *account.Account) *account.AcctProto {
	return &account.AcctProto{
		Address:  acct.GetAddress(),
		Name:     acct.GetName(),
		Nonce:    acct.GetNonce(),
		XBalance: acct.GetBalance().Bytes(),
		XCode:    acct.GetCode(),
	}
}

func FromProto(pm *account.AcctProto) *account.Account {
	return &account.Account{
		Address: pm.Address,
		Name:    pm.Name,
		Nonce:   pm.Nonce,
		Balance: new(big.Int).SetBytes(pm.XBalance),
		Code:    pm.XCode,
	}
}
