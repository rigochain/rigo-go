package account

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
)

func (ctrler *AccountCtrler) Query(qd *types.QueryData) (json.RawMessage, xerrors.XError) {
	switch qd.Command {
	case types.QUERY_ACCOUNT:
		addr := types.Address(qd.Params)
		if acct := ctrler.ReadAccount(addr); acct == nil {
			return nil, xerrors.ErrNotFoundAccount
		} else if raw, err := json.Marshal(acct.(*Account)); err != nil {
			return nil, xerrors.ErrQuery.With(err)
		} else {
			return raw, nil
		}
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
