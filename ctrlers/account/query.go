package account

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *AcctCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	addr := types.Address(req.Data)
	if acct := ctrler.ReadAccount(addr); acct == nil {
		return nil, xerrors.ErrNotFoundAccount
	} else if raw, err := tmjson.Marshal(acct); err != nil {
		return nil, xerrors.ErrQuery.With(err)
	} else {
		return raw, nil
	}
}
