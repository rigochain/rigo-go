package gov

import (
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *GovCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	txhash := req.Data
	switch req.Path {
	case "proposals":
		if txhash == nil || len(txhash) == 0 {
			if propos, xerr := ctrler.GetProposals(); xerr != nil {
				if xerr == xerrors.ErrNotFoundResult {
					return nil, xerrors.ErrNotFoundProposal
				}
				return nil, xerr
			} else if v, err := tmjson.Marshal(propos); err != nil {
				return nil, xerrors.ErrQuery.Wrap(err)
			} else {
				return v, nil
			}
		} else {
			if propo, xerr := ctrler.ReadProposals(txhash); xerr != nil {
				if xerr == xerrors.ErrNotFoundResult {
					return nil, xerrors.ErrNotFoundProposal
				}
				return nil, xerr
			} else if v, err := tmjson.Marshal(propo); err != nil {
				return nil, xerrors.ErrQuery.Wrap(err)
			} else {
				return v, nil
			}
		}
	case "rule":
		if bz, err := tmjson.Marshal(&ctrler.GovRule); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return bz, nil
		}
	}

	return nil, nil
}
