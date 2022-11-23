package gov

import (
	"encoding/json"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *GovCtrler) Query(req abcitypes.RequestQuery) (json.RawMessage, xerrors.XError) {
	txhash := req.Data
	if txhash == nil || len(txhash) == 0 {
		propos := ctrler.GetProposals()
		if propos == nil {
			return nil, xerrors.ErrNotFoundProposal
		} else if v, err := json.Marshal(propos); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return v, nil
		}
	} else {
		propo := ctrler.FindProposals(txhash)
		if propo == nil {
			return nil, xerrors.ErrNotFoundProposal
		} else if v, err := propo.Encode(); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return v, nil
		}
	}
}
