package gov

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
	"sync"
)

type GovCtrler struct {
	govDB tmdb.DB

	rules types.IGovRules

	allProposals     map[[32]byte]types.IProposable
	updatedProposals []types.IProposable

	logger log.Logger
	mtx    sync.RWMutex
}

func NewGovCtrler(dbDir string, logger log.Logger) (*GovCtrler, error) {
	govDB, err := tmdb.NewDB("gov", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	iter, err := govDB.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	allProposals := make(map[[32]byte]types.IProposable)
	for ; iter.Valid(); iter.Next() {
		k := iter.Key()
		v := iter.Value()
		p := &GovRulesProposal{}
		if err := tmjson.Unmarshal(v, p); err != nil {
			return nil, err
		}
		allProposals[types.HexBytes(k).Array32()] = p
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return &GovCtrler{govDB: govDB, allProposals: allProposals, logger: logger}, nil
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

var _ trxs.ITrxHandler = (*GovCtrler)(nil)
var _ types.ILedgerCtrler = (*GovCtrler)(nil)

func (ctrler *GovCtrler) Validate(context *trxs.TrxContext) error {
	// todo: only validators can propose
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

		if _, ok := ctrler.allProposals[ctx.TxHash.Array32()]; ok {
			return xerrors.New("there is already the txhash in allProposals")
		}

		proposal := &GovRulesProposal{
			TxHash:            ctx.TxHash,
			StartVotingHeight: ctx.Height + 1,
			LastVotingHeight:  ctx.Height + txpayload.VotingBlocks,
			ApplyingHeight:    ctx.Height + txpayload.VotingBlocks + ctx.GovRules.GetLazyApplyingBlocks(),
			MajorityPower:     ctx.StakeCtrler.GetTotalPower() * int64(2) / int64(3),
			Votes:             make(map[string]int64),
			Rules:             rules,
		}
		ctrler.updatedProposals = append(ctrler.updatedProposals, proposal)
		ctrler.allProposals[ctx.TxHash.Array32()] = proposal
	}
	return nil
}

func (ctrler *GovCtrler) applyVote(ctx *trxs.TrxContext) error {
	votingPayload := ctx.Tx.Payload.(*trxs.TrxPayloadVoting)
	proposal, ok := ctrler.allProposals[votingPayload.TxHash.Array32()]
	if !ok {
		return xerrors.New("not found proposal")
	}
	votes := proposal.GetVotesOf(ctx.Tx.From, votingPayload.Choice)
	if votes > 0 {
		return xerrors.New("already voted")
	}

	proposal.DoVote(ctx.Tx.From, ctx.StakeCtrler.GetTotalPowerOf(ctx.Tx.From))
	ctrler.updatedProposals = append(ctrler.updatedProposals, proposal)
	return nil
}

func (ctrler *GovCtrler) Commit() ([]byte, int64, error) {
	for _, proposal := range ctrler.updatedProposals {
		if bz, err := proposal.Encode(); err != nil {
			return nil, -1, err
		} else if err := ctrler.govDB.Set(proposal.ID(), bz); err != nil {
			return nil, -1, err
		}
	}
	ctrler.updatedProposals = nil
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
