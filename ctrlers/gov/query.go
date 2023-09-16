package gov

import (
	"github.com/rigochain/rigo-go/ctrlers/gov/proposal"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

func (ctrler *GovCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	txhash := req.Data
	switch req.Path {
	case "proposal":
		atledger, xerr := ctrler.proposalLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}

		if txhash == nil || len(txhash) == 0 {
			var proposals []*proposal.GovProposal
			if xerr := atledger.IterateReadAllItems(func(prop *proposal.GovProposal) xerrors.XError {
				proposals = append(proposals, prop)
				return nil
			}); xerr != nil {
				return nil, xerrors.ErrQuery.Wrap(xerr)
			}

			v, err := tmjson.Marshal(proposals)
			if err != nil {
				return nil, xerrors.ErrQuery.Wrap(err)
			}
			return v, nil
		} else {
			prop, xerr := atledger.Read(ledger.ToLedgerKey(txhash))
			if xerr != nil {
				return nil, xerrors.ErrQuery.Wrap(xerr)
			}

			v, err := tmjson.Marshal(prop)
			if err != nil {
				return nil, xerrors.ErrQuery.Wrap(err)
			}
			return v, nil
		}
	case "gov_params":
		atledger, xerr := ctrler.paramsLedger.ImmutableLedgerAt(req.Height, 0)
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		govParams, xerr := atledger.Read(ledger.ToLedgerKey(bytes.ZeroBytes(32)))
		if xerr != nil {
			return nil, xerrors.ErrQuery.Wrap(xerr)
		}
		bz, err := tmjson.Marshal(govParams)
		if err != nil {
			return nil, xerrors.ErrQuery.Wrap(err)
		}
		return bz, nil
	}

	return nil, nil
}
