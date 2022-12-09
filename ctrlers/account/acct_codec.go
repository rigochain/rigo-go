package account

import (
	types2 "github.com/kysee/arcanus/ctrlers/types"
	"google.golang.org/protobuf/proto"
	"math/big"
)

func EncodeAccount(acct *types2.Account) ([]byte, error) {
	return proto.Marshal(ToProto(acct))
}

func DecodeAccount(bz []byte) (*types2.Account, error) {
	pm := &types2.AcctProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return nil, err
	}
	return FromProto(pm), nil
}

func ToProto(acct *types2.Account) *types2.AcctProto {
	return &types2.AcctProto{
		Address:  acct.GetAddress(),
		Name:     acct.GetName(),
		Nonce:    acct.GetNonce(),
		XBalance: acct.GetBalance().Bytes(),
		XCode:    acct.GetCode(),
	}
}

func FromProto(pm *types2.AcctProto) *types2.Account {
	return &types2.Account{
		Address: pm.Address,
		Name:    pm.Name,
		Nonce:   pm.Nonce,
		Balance: new(big.Int).SetBytes(pm.XBalance),
		Code:    pm.XCode,
	}
}
