package stake

import (
	"encoding/json"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (ctrler *StakeCtrler) Query(req abcitypes.RequestQuery) (json.RawMessage, xerrors.XError) {
	addr := account.Address(req.Data)
	if staker := ctrler.FindDelegatee(addr); staker == nil {
		return nil, xerrors.ErrNotFoundStaker
	} else if v, err := json.Marshal(staker); err != nil {
		return nil, xerrors.ErrQuery.With(err)
	} else {
		return v, nil
	}
}
