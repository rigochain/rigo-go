package gov

import (
	"github.com/kysee/arcanus/ctrlers/gov/proposals"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
	"sync"
)

type GovCtrler struct {
	govDB tmdb.DB

	rules types.IGovRules

	proposals map[[32]byte]types.IProposable

	logger log.Logger
	mtx    sync.RWMutex
}

var _ trxs.ITrxHandler = (*GovCtrler)(nil)
var _ types.ILedgerCtrler = (*GovCtrler)(nil)

func (ctrler *GovCtrler) Validate(context *trxs.TrxContext) error {
	return nil
}

func (ctrler *GovCtrler) Apply(ctx *trxs.TrxContext) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch ctx.Tx.GetType() {
	case trxs.TRX_PROPOSAL:
		return ctrler.applyProposal(ctx)
	case trxs.TRX_VOTING:
		return ctrler.applyVote(ctx)
	default:
		return xerrors.New("unknown transaction type")
	}
}

func (ctrler *GovCtrler) applyProposal(ctx *trxs.TrxContext) error {
	if txpayload, ok := ctx.Tx.Payload.(*trxs.TrxPayloadProposal); !ok {
		return xerrors.New("unknown transaction payload")
	} else {
		var rules []*GovRules
		for _, opt := range txpayload.Options {
			if r, err := DecodeGovRules(opt); err != nil {
				return err
			} else {
				rules = append(rules, r)
			}

		}

		if _, ok := ctrler.proposals[ctx.TxHash.Array32()]; ok {
			return xerrors.New("there is already the txhash in proposals")
		}

		ctrler.proposals[ctx.TxHash.Array32()] = &proposals.GovRulesProposal{
			StartVotingHeight: ctx.Height + 1,
			LastVotingHeight:  ctx.Height + txpayload.VotingBlocks,
			ApplyingHeight:    ctx.Height + txpayload.VotingBlocks + ctx.GovRules.GetLazyApplyingBlocks(),
			MajorityPower:     ctx.StakeCtrler.GetTotalPower() * int64(2) / int64(3),
			Rules:             rules,
		}
	}
	return nil
}

func (ctrler *GovCtrler) applyVote(ctx *trxs.TrxContext) error {
	return nil
}

func (ctrler *GovCtrler) Commit() ([]byte, int64, error) {
	return nil, 0, nil
}

func (ctrler *GovCtrler) Close() error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.govDB != nil {
		if err := ctrler.govDB.Close(); err != nil {
			return nil
		}
	}

	ctrler.govDB = nil
	return nil
}

func NewGovCtrler(dbDir string, logger log.Logger) (*GovCtrler, error) {
	// todo: use govDB to save governance rules proposal, voting actions, etc.
	govDB, err := tmdb.NewDB("gov", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	return &GovCtrler{govDB: govDB, logger: logger}, nil
}

func (ctrler *GovCtrler) SetRules(rules types.IGovRules) {
	ctrler.rules = rules
}

func (ctrler *GovCtrler) GetRules() types.IGovRules {
	return ctrler.rules
}

func (ctrler *GovCtrler) ImportRules(cb func() []byte) error {
	bz := cb()
	if bz == nil {
		return xerrors.New("rule blob is nil")
	} else if rules, err := DecodeGovRules(bz); err != nil {
		return xerrors.NewFrom(err)
	} else {
		ctrler.rules = rules
	}
	return nil
}
