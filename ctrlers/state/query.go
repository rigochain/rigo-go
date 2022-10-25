package state

import (
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/json"
)

func (ctrler *ChainCtrler) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	response := abcitypes.ResponseQuery{
		Code: abcitypes.CodeTypeOK,
		Key:  req.Data,
	}

	qd, xerr := types.DecodeQueryData(req.Data)
	if xerr != nil {
		response.Code = xerr.(xerrors.XError).Code()
		response.Log = xerr.Error()
		return response
	}

	switch qd.Command {
	case types.QUERY_ACCOUNT:
		addr := types.Address(qd.Params)
		acct := ctrler.acctCtrler.ReadAccount(addr)
		if acct == nil {
			response.Code = xerrors.ErrNotFoundAccount.Code()
			response.Log = xerrors.ErrNotFoundAccount.Error()
		} else if v, err := account.EncodeAccount(acct); err != nil {
			response.Code = xerrors.ErrQuery.Code()
			response.Log = xerrors.ErrQuery.With(err).Error()
		} else {
			response.Value = v
		}
	case types.QUERY_STAKES:
		addr := types.Address(qd.Params)
		if staker := ctrler.stakeCtrler.FindDelegatee(addr); staker == nil {
			response.Code = xerrors.ErrNotFoundStaker.Code()
			response.Log = xerrors.ErrNotFoundStaker.Error()
		} else if v, err := json.Marshal(staker); err != nil {
			response.Code = xerrors.ErrQuery.Code()
			response.Log = xerrors.ErrQuery.With(err).Error()
		} else {
			response.Value = v
		}
	default:
		response.Code = xerrors.ErrInvalidQueryCmd.Code()
		response.Log = xerrors.ErrInvalidQueryCmd.Error()
	}

	return response
}
