package gov

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
)

func (ctrler *GovCtrler) Query(qd *types.QueryData) ([]byte, xerrors.XError) {
	switch qd.Command {
	case types.QUERY_PROPOSALS:
		return nil, nil
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
