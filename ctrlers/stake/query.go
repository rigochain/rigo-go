package stake

import (
	"github.com/kysee/arcanus/ledger"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/bytes"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *StakeCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	switch req.Path {
	case "stakes":
		addr := types.Address(req.Data)
		var stakes []*Stake
		if err := ctrler.delegateeLedger.IterateAllItems(func(d *Delegatee) xerrors.XError {
			for _, s0 := range d.Stakes {
				if bytes.Compare(s0.From, addr) == 0 {
					stakes = append(stakes, s0)
				}
			}
			return nil
		}); err != nil {
			return nil, xerrors.ErrQuery.With(err)
		} else if bz, err := tmjson.Marshal(stakes); err != nil {
			return nil, xerrors.ErrQuery.With(err)
		} else {
			return bz, nil
		}
	case "delegatee":
		addr := types.Address(req.Data)
		if delegatee, xerr := ctrler.delegateeLedger.Read(ledger.ToLedgerKey(addr)); xerr != nil {
			if xerr == xerrors.ErrNotFoundResult {
				return nil, xerrors.ErrNotFoundDelegatee
			}
			return nil, xerr
		} else if v, err := tmjson.Marshal(delegatee); err != nil {
			return nil, xerrors.ErrQuery.With(err)
		} else {
			return v, nil
		}
	default:
		return nil, xerrors.New("unknown query path")
	}
}
