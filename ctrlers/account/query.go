package account

import (
	"encoding/json"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *AccountCtrler) Query(req abcitypes.RequestQuery) (json.RawMessage, xerrors.XError) {
	addr := account.Address(req.Data)
	if acct := ctrler.ReadAccount(addr); acct == nil {
		return nil, xerrors.ErrNotFoundAccount
	} else if raw, err := json.Marshal(acct); err != nil {
		return nil, xerrors.ErrQuery.With(err)
	} else {
		return raw, nil
	}
}
