package account

import (
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
	"math/big"
)

func EncodeAccount(acct types.IAccount) ([]byte, error) {
	return proto.Marshal(ToProto(acct))
}

func DecodeAccount(bz []byte) (*Account, error) {
	pm := &AcctProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return nil, err
	}
	return FromProto(pm), nil
}

func ToProto(acct types.IAccount) *AcctProto {
	return &AcctProto{
		Address:  acct.GetAddress(),
		Name:     acct.GetName(),
		Nonce:    acct.GetNonce(),
		XBalance: acct.GetBalance().Bytes(),
		XCode:    acct.GetCode(),
	}
}

func FromProto(pm *AcctProto) *Account {
	return &Account{
		Address: pm.Address,
		Name:    pm.Name,
		Nonce:   pm.Nonce,
		Balance: new(big.Int).SetBytes(pm.XBalance),
		Code:    pm.XCode,
	}
}
