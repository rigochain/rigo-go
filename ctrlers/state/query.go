package state

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *ChainCtrler) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	response := abcitypes.ResponseQuery{
		Code: abcitypes.CodeTypeOK,
		Key:  req.Data,
	}
	if qd, xerr := types.DecodeQueryData(req.Data); xerr != nil {
		response.Code = xerr.Code()
		response.Log = xerr.Error()
	} else {
		switch qd.Command {
		case types.QUERY_ACCOUNT:
			response.Value, xerr = ctrler.acctCtrler.Query(qd)
		case types.QUERY_STAKES:
			response.Value, xerr = ctrler.stakeCtrler.Query(qd)
		case types.QUERY_PROPOSALS:
			response.Value, xerr = ctrler.govCtrler.Query(qd)
		default:
			response.Value, xerr = nil, xerrors.ErrInvalidQueryCmd
		}
	}
	return response
}
