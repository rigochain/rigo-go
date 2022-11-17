package stake

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/xerrors"
)

func (ctrler *StakeCtrler) Query(qd *types.QueryData) (json.RawMessage, xerrors.XError) {

	switch qd.Command {
	case types.QUERY_STAKES:
		addr := account.Address(qd.Params)
		if staker := ctrler.FindDelegatee(addr); staker == nil {
			return nil, xerrors.ErrNotFoundStaker
		} else if v, err := json.Marshal(staker); err != nil {
			return nil, xerrors.ErrQuery.With(err)
		} else {
			return v, nil
		}
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
