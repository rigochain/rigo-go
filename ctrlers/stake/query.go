package stake

import (
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *StakeCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	switch req.Path {
	case "reward":
		atledger, xerr := ctrler.rewardLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		rwd, xerr := atledger.Read(ledger.ToLedgerKey(req.Data))
		if rwd == nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		bz, err := tmjson.Marshal(rwd)
		if err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		}
		return bz, nil
	case "stakes":
		atledger, xerr := ctrler.delegateeLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}

		var stakes []*Stake
		if err := atledger.IterateReadAllItems(func(d *Delegatee) xerrors.XError {
			for _, s0 := range d.Stakes {
				if bytes.Compare(s0.From, types.Address(req.Data)) == 0 {
					stakes = append(stakes, s0)
				}
			}
			return nil
		}); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else if bz, err := tmjson.Marshal(stakes); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return bz, nil
		}
	case "delegatee":
		atledger, xerr := ctrler.delegateeLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}

		if delegatee, xerr := atledger.Read(ledger.ToLedgerKey(req.Data)); xerr != nil {
			if xerr == xerrors.ErrNotFoundResult {
				return nil, xerrors.ErrQuery.Wrap(xerrors.ErrNotFoundDelegatee)
			}
			return nil, xerr
		} else if v, err := tmjson.Marshal(delegatee); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return v, nil
		}
	default:
		return nil, xerrors.ErrQuery.Wrapf("unknown query path")
	}
}
