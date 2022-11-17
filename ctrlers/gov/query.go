package gov

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
)

func (ctrler *GovCtrler) Query(qd *types.QueryData) (json.RawMessage, xerrors.XError) {
	switch qd.Command {
	case types.QUERY_PROPOSALS:
		txhash := qd.Params
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
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
