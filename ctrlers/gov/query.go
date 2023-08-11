package gov

import (
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *GovCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	txhash := req.Data
	switch req.Path {
	case "proposals":
		if txhash == nil || len(txhash) == 0 {
			if propos, xerr := ctrler.RealAllProposals(); xerr != nil {
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
			if propo, xerr := ctrler.ReadProposal(txhash); xerr != nil {
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
		atledger, xerr := ctrler.ruleLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		govParams, xerr := atledger.Read(ledger.ToLedgerKey(bytes.ZeroBytes(32)))
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		if bz, err := tmjson.Marshal(govParams); err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		} else {
			return bz, nil
		}
	}

	return nil, nil
}
