package node

import (
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *ArcanusApp) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	response := abcitypes.ResponseQuery{
		Code: abcitypes.CodeTypeOK,
		Key:  req.Data,
	}

	var xerr xerrors.XError

	switch req.Path {
	case "account":
		response.Value, xerr = ctrler.acctCtrler.Query(req)
	case "stakes", "delegatee":
		response.Value, xerr = ctrler.stakeCtrler.Query(req)
	case "proposals", "rule":
		response.Value, xerr = ctrler.govCtrler.Query(req)
	default:
		response.Value, xerr = nil, xerrors.ErrInvalidQueryPath
	}

	if xerr != nil {
		response.Code = xerr.Code()
		response.Log = xerr.Error()
	}
	return response
}
