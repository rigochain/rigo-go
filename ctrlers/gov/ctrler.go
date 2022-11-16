package gov

import (
	"fmt"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"math/big"
	"sync"
)

type GovCtrler struct {
	govDB tmdb.DB

	govRule *GovRule

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
		p := &GovRuleProposal{}
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

func (ctrler *GovCtrler) SetRules(r *GovRule) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.govRule = r
}

func (ctrler *GovCtrler) GetRules() *GovRule {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.govRule
}

func (ctrler *GovCtrler) ImportRules(cb func() []byte) error {

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	bz := cb()
	if bz == nil {
		return xerrors.New("rule blob is nil")
	} else if rule, err := DecodeGovRule(bz); err != nil {
		return xerrors.NewFrom(err)
	} else {
		ctrler.govRule = rule
	}
	return nil
}

var _ trxs.ITrxHandler = (*GovCtrler)(nil)
var _ types.ILedgerHandler = (*GovCtrler)(nil)

func (ctrler *GovCtrler) Validate(ctx *trxs.TrxContext) error {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	switch ctx.Tx.GetType() {
	case trxs.TRX_PROPOSAL, trxs.TRX_VOTING:
		if ctx.StakeHandler.IsValidator(ctx.Tx.From) == false {
			return xerrors.New("wrong proposer")
		}
	default:
		return xerrors.New("unknown transaction type")
	}

	return nil
}

func (ctrler *GovCtrler) Execute(ctx *trxs.TrxContext) error {
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
		var rules []*GovRule
		for _, opt := range txpayload.Options {
			if r, err := DecodeGovRule(opt); err != nil {
				return err
			} else {
				rules = append(rules, r)
			}

		}

		if _, ok := ctrler.allProposals[ctx.TxHash.Array32()]; ok {
			return xerrors.New("there is already the txhash in allProposals")
		}

		proposal := &GovRuleProposal{
			TxHash:            ctx.TxHash,
			StartVotingHeight: ctx.Height + 1,
			LastVotingHeight:  ctx.Height + txpayload.VotingBlocks,
			ApplyingHeight:    ctx.Height + txpayload.VotingBlocks + ctrler.govRule.LazyApplyingBlocks,
			MajorityPower:     ctx.StakeHandler.GetTotalPower() * int64(2) / int64(3),
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

	proposal.DoVote(ctx.Tx.From, ctx.StakeHandler.GetTotalPowerOf(ctx.Tx.From))
	ctrler.updatedProposals = append(ctrler.updatedProposals, proposal)
	return nil
}

func (ctrler *GovCtrler) Commit() ([]byte, int64, error) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

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

// implements IGovRuleHandler
func (ctrler *GovCtrler) GetMaxValidatorCount() int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.govRule.MaxValidatorCnt
}

// MaxStakeAmount means the max of amount which could be deposited.
// tmtypes.MaxTotalVotingPower = int64(math.MaxInt64) / 8
// When the type of voting power is `int64`and VP:XCO = 1:1,
// the MAXSTAKEsau becomes `int64(math.MaxInt64) / 8 * 10^18` (~= 922경 XCO)
func (ctrler *GovCtrler) MaxStakeAmount() *big.Int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return new(big.Int).Mul(big.NewInt(tmtypes.MaxTotalVotingPower), ctrler.govRule.AmountPerPower)
}

func (ctrler *GovCtrler) MaxTotalPower() int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return tmtypes.MaxTotalVotingPower
}

func (ctrler *GovCtrler) AmountToPower(amt *big.Int) int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, ctrler.govRule.AmountPerPower)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (ctrler *GovCtrler) PowerToAmount(power int64) *big.Int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), ctrler.govRule.AmountPerPower)
}

func (ctrler *GovCtrler) PowerToReward(power int64) *big.Int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), ctrler.govRule.RewardPerPower)
}

func (ctrler *GovCtrler) GetLazyRewardBlocks() int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.govRule.LazyRewardBlocks
}

func (ctrler *GovCtrler) GetLazyApplyingBlocks() int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.govRule.LazyApplyingBlocks
}

var _ types.IGovRuleHandler = (*GovCtrler)(nil)
