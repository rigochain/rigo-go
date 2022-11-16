package account

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
)

func (ctrler *AccountCtrler) Query(qd *types.QueryData) ([]byte, xerrors.XError) {
	switch qd.Command {
	case types.QUERY_ACCOUNT:
		addr := types.Address(qd.Params)
		acct := ctrler.ReadAccount(addr)
		if acct == nil {
			return nil, xerrors.ErrNotFoundAccount
		} else if v, err := EncodeAccount(acct); err != nil {
			return nil, xerrors.ErrQuery
		} else {
			return v, nil
		}
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
