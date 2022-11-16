package stake

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *StakeCtrler) Query(qd *types.QueryData) ([]byte, xerrors.XError) {

	switch qd.Command {
	case types.QUERY_STAKES:
		addr := types.Address(qd.Params)
		if staker := ctrler.FindDelegatee(addr); staker == nil {
			return nil, xerrors.ErrNotFoundStaker
		} else if v, err := tmjson.Marshal(staker); err != nil {
			return nil, xerrors.ErrQuery
		} else {
			return v, nil
		}
	default:
		return nil, xerrors.ErrInvalidQueryCmd
	}
}
