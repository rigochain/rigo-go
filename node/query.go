package node

import (
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *RigoApp) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	response := abcitypes.ResponseQuery{
		Code:   abcitypes.CodeTypeOK,
		Key:    req.Data,
		Height: req.Height,
	}

	var xerr xerrors.XError

	switch req.Path {
	case "account":
		response.Value, xerr = ctrler.acctCtrler.Query(req)
	case "stakes", "delegatee", "reward":
		response.Value, xerr = ctrler.stakeCtrler.Query(req)
	case "proposals", "gov_params":
		response.Value, xerr = ctrler.govCtrler.Query(req)
	case "vm_call":
		response.Value, xerr = ctrler.vmCtrler.Query(req)
	default:
		response.Value, xerr = nil, xerrors.ErrInvalidQueryPath
	}

	if xerr != nil {
		response.Code = xerr.Code()
		response.Log = xerr.Error()
	}

	return response
}
