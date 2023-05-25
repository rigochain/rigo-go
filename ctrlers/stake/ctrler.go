package stake

import (
	"fmt"
	"github.com/holiman/uint256"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"sort"
	"strconv"
	"sync"
)

type InitStake struct {
	PubKeys bytes.HexBytes
	Stakes  []*Stake
}

type StakeCtrler struct {
	lastValidators  DelegateeArray
	delegateeLedger ledger.IFinalityLedger[*Delegatee]
	frozenLedger    ledger.IFinalityLedger[*Stake]

	govParams ctrlertypes.IGovHandler

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewStakeCtrler(config *cfg.Config, govHandler ctrlertypes.IGovHandler, logger tmlog.Logger) (*StakeCtrler, error) {
	//totalPowerOfDelegatees, totalPowerOfStakes := int64(0), int64(0)
	newDelegateeProvider := func() *Delegatee { return &Delegatee{} }
	newStakeProvider := func() *Stake { return &Stake{} }

	// for all delegatees
	delegateeLedger, xerr := ledger.NewFinalityLedger[*Delegatee]("delegatees", config.DBDir(), 128, newDelegateeProvider)
	if xerr != nil {
		return nil, xerr
	}

	frozenLedger, xerr := ledger.NewFinalityLedger[*Stake]("frozen", config.DBDir(), 128, newStakeProvider)
	if xerr != nil {
		return nil, xerr
	}

	ret := &StakeCtrler{
		delegateeLedger: delegateeLedger,
		//stakeLedger:     stakeLedger,
		frozenLedger: frozenLedger,
		govParams:    govHandler,
		logger:       logger.With("module", "rigo_StakeCtrler"),
	}
	return ret, nil
}

func (ctrler *StakeCtrler) InitLedger(req interface{}) xerrors.XError {
	// init validators
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	initStakes, ok := req.([]*InitStake)
	if !ok {
		return xerrors.ErrInitChain.Wrapf("wrong parameter: StakeCtrler::InitLedger() requires []*InitStake")
	}

	for _, initS0 := range initStakes {
		for _, s0 := range initS0.Stakes {
			d := NewDelegatee(s0.To, initS0.PubKeys)
			if xerr := d.AddStake(s0); xerr != nil {
				return xerr
			}
			if xerr := ctrler.delegateeLedger.SetFinality(d); xerr != nil {
				return xerr
			}
		}
	}

	return nil
}

// BeginBlock are called in RigoApp::BeginBlock
func (ctrler *StakeCtrler) BeginBlock(blockCtx *ctrlertypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	var evts []abcitypes.Event

	byzantines := blockCtx.BlockInfo().ByzantineValidators
	if byzantines != nil {
		ctrler.logger.Debug("Byzantine validators is found", "count", len(byzantines))
		for _, evi := range byzantines {
			if slashed, xerr := ctrler.doPunish(
				&evi, blockCtx.GovHandler.SlashRatio(),
				ctrlertypes.AmountPerPower(),
				blockCtx.GovHandler.RewardPerPower()); xerr != nil {
				ctrler.logger.Error("Error when punishing",
					"byzantine", types.Address(evi.Validator.Address),
					"evidenceType", abcitypes.EvidenceType_name[int32(evi.Type)])
			} else {
				_addr := types.Address(evi.Validator.Address).String()
				_type := abcitypes.EvidenceType_name[int32(evi.Type)]
				_power0 := strconv.FormatInt(evi.Validator.Power, 10)
				_slashed := strconv.FormatInt(slashed, 10)
				evts = append(evts, abcitypes.Event{
					Type: "punishment",
					Attributes: []abcitypes.EventAttribute{
						{Key: []byte("byzantine"), Value: []byte(_addr), Index: true},
						{Key: []byte("type"), Value: []byte(_type), Index: false},
						{Key: []byte("power"), Value: []byte(_power0), Index: false},
						{Key: []byte("slashed"), Value: []byte(_slashed), Index: false},
					},
				})
			}
		}
	}

	return evts, nil
}

func (ctrler *StakeCtrler) DoPunish(evi *abcitypes.Evidence, slashRatio int64, amtPerPower *uint256.Int, rwdPerPower int64) (int64, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.doPunish(evi, slashRatio, amtPerPower, rwdPerPower)
}

func (ctrler *StakeCtrler) doPunish(evi *abcitypes.Evidence, slashRatio int64, amtPerPower *uint256.Int, rwdPerPower int64) (int64, xerrors.XError) {
	delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(evi.Validator.Address))
	if xerr != nil {
		return 0, xerr
	}

	// Punish the delegators as well as validator. issue #51
	slashed := delegatee.DoSlash(slashRatio, amtPerPower, rwdPerPower, true)
	_ = ctrler.delegateeLedger.SetFinality(delegatee)

	return slashed, nil
}

func (ctrler *StakeCtrler) ValidateTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_STAKING:
		q, r := new(uint256.Int).DivMod(ctx.Tx.Amount, ctrlertypes.AmountPerPower(), new(uint256.Int))
		// `ctx.Tx.Amount` MUST be greater than or equal to `ctrler.govHelper.AmountPerPower()`
		//    ==> q.Sign() > 0
		if q.Sign() <= 0 {
			return xerrors.ErrInvalidTrx.Wrapf("wrong amount: it should be greater than %v", ctrlertypes.AmountPerPower())
		}
		// `ctx.Tx.Amount` MUST be multiple to `ctrler.govHelper.AmountPerPower()`
		//    ==> r.Sign() == 0
		if r.Sign() != 0 {
			return xerrors.ErrInvalidTrx.Wrapf("wrong amount: it should be multiple of %v", ctrlertypes.AmountPerPower())
		}

		if bytes.Compare(ctx.Tx.From, ctx.Tx.To) == 0 {
			// isseu #59
			// self-staking to become a validator
			// check MinValidatorStake

			getDelegatee := ctrler.delegateeLedger.Get
			if ctx.Exec {
				getDelegatee = ctrler.delegateeLedger.GetFinality
			}

			expectedStake := ctx.Tx.Amount
			if delegatee, _ := getDelegatee(ledger.ToLedgerKey(ctx.Tx.From)); delegatee != nil {
				expectedStake = new(uint256.Int).Add(ctx.Tx.Amount, delegatee.SelfAmount)
			}

			if expectedStake.Lt(ctrler.govParams.MinValidatorStake()) {
				return xerrors.ErrInvalidTrx.Wrapf("too small stake to become validator: a minimum is %v", ctrler.govParams.MinValidatorStake())
			}

		}

	case ctrlertypes.TRX_UNSTAKING:
	default:
		return xerrors.ErrUnknownTrxType
	}

	return nil
}

func (ctrler *StakeCtrler) ExecuteTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	// executing staking and un-staking txs

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_STAKING:
		return ctrler.execStaking(ctx)
	case ctrlertypes.TRX_UNSTAKING:
		return ctrler.execUnstaking(ctx)
	default:
		return xerrors.ErrUnknownTrxType
	}
}

func (ctrler *StakeCtrler) execStaking(ctx *ctrlertypes.TrxContext) xerrors.XError {
	getDelegatee := ctrler.delegateeLedger.Get
	setUpdateDelegatee := ctrler.delegateeLedger.Set
	if ctx.Exec {
		getDelegatee = ctrler.delegateeLedger.GetFinality
		setUpdateDelegatee = ctrler.delegateeLedger.SetFinality
	}

	delegatee, xerr := getDelegatee(ledger.ToLedgerKey(ctx.Tx.To))
	if xerr != nil && xerr != xerrors.ErrNotFoundResult {
		return xerr
	}

	if delegatee == nil && bytes.Compare(ctx.Tx.From, ctx.Tx.To) == 0 {
		// add new delegatee
		delegatee = NewDelegatee(ctx.Tx.From, ctx.SenderPubKey)
	} else if delegatee == nil {
		// there is no delegatee whose address is ctx.Tx.To
		return xerrors.ErrNotFoundDelegatee
	}

	// create stake and delegate it to delegatee
	// the block reward for this stake will be started at ctx.Height + 1. (issue #29)
	s0 := NewStakeWithAmount(ctx.Tx.From, ctx.Tx.To, ctx.Tx.Amount, ctx.Height+1, ctx.TxHash)
	if !s0.IsSelfStake() {
		// it's delegating. check minSelfStakeRatio
		selfRatio := delegatee.SelfStakeRatio(s0.Power)
		if selfRatio < ctx.GovHandler.MinSelfStakeRatio() {
			return xerrors.From(fmt.Errorf("not enough self power - validator: %v, self power: %v", delegatee.Addr, delegatee.GetSelfPower()))
		}
	}

	if xerr := delegatee.AddStake(s0); xerr != nil {
		return xerr
	}
	if xerr := setUpdateDelegatee(delegatee); xerr != nil {
		return xerr
	}

	return nil
}

func (ctrler *StakeCtrler) execUnstaking(ctx *ctrlertypes.TrxContext) xerrors.XError {
	getDelegatee := ctrler.delegateeLedger.Get
	setUpdateDelegatee := ctrler.delegateeLedger.Set
	delDelegatee := ctrler.delegateeLedger.Del
	setUpdateFrozen := ctrler.frozenLedger.Set
	if ctx.Exec {
		getDelegatee = ctrler.delegateeLedger.GetFinality
		setUpdateDelegatee = ctrler.delegateeLedger.SetFinality
		delDelegatee = ctrler.delegateeLedger.DelFinality
		setUpdateFrozen = ctrler.frozenLedger.SetFinality
	}

	// find delegatee
	delegatee, xerr := getDelegatee(ledger.ToLedgerKey(ctx.Tx.To))
	if xerr != nil {
		return xerr
	}

	// delete the stake from a delegatee
	txhash := ctx.Tx.Payload.(*ctrlertypes.TrxPayloadUnstaking).TxHash
	if txhash == nil || len(txhash) != 32 {
		return xerrors.ErrInvalidTrxPayloadParams
	}

	_, s0 := delegatee.FindStake(txhash)
	if s0 == nil {
		return xerrors.ErrNotFoundStake
	}

	// issue #43
	// check that tx's sender is stake's owner
	if ctx.Tx.From.Compare(s0.From) != 0 {
		return xerrors.ErrNotFoundStake.Wrapf("you not stake owner")
	}

	_ = delegatee.DelStake(txhash)

	s0.RefundHeight = ctx.Height + ctx.GovHandler.LazyRewardBlocks()
	_ = setUpdateFrozen(s0) // add s0 to frozen ledger

	if delegatee.SelfPower == 0 {
		stakes := delegatee.DelAllStakes()
		for _, _s0 := range stakes {
			_s0.RefundHeight = ctx.Height + ctx.GovHandler.LazyRewardBlocks()
			_ = setUpdateFrozen(_s0) // add s0 to frozen ledger
		}
	}

	if delegatee.TotalPower == 0 {
		// this changed delegate will be committed at Commit()
		if _, xerr := delDelegatee(delegatee.Key()); xerr != nil {
			return xerr
		}
	} else {
		// this changed delegate will be committed at Commit()
		if xerr := setUpdateDelegatee(delegatee); xerr != nil {
			return xerr
		}
	}

	return nil
}

func (ctrler *StakeCtrler) EndBlock(ctx *ctrlertypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// reward validators for all blocks (including empty blocks): issue #54
	if xerr := ctrler.doReward(ctx.BlockInfo().Header.Height, ctx.BlockInfo().LastCommitInfo.Votes); xerr != nil {
		return nil, xerr
	}

	if xerr := ctrler.unfreezingStakes(ctx.Height(), ctx.AcctHandler); xerr != nil {
		return nil, xerr
	}

	ctx.SetValUpdates(ctrler.updateValidators(int(ctx.GovHandler.MaxValidatorCnt())))

	return nil, nil
}
func (ctrler *StakeCtrler) DoReward(height int64, votes []abcitypes.VoteInfo) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.doReward(height, votes)
}

// the following functions are called in EndBlock()
//

func (ctrler *StakeCtrler) doReward(height int64, votes []abcitypes.VoteInfo) xerrors.XError {
	// doReward() rewards to NOT ctrler.lastValidators BUT `votes`.
	// `votes` is from `RequestBeginBlock.LastCommitInfo` which is set of validators of previous block.
	for _, vote := range votes {
		if vote.SignedLastBlock == false {
			ctrler.logger.Debug("Validator didn't sign the last block", "address", types.Address(vote.Validator.Address), "power", vote.Validator.Power)
			continue
		}

		delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(vote.Validator.Address))
		if xerr != nil {
			ctrler.logger.Error("Not found validator", "error", xerr, "address", types.Address(vote.Validator.Address), "power", vote.Validator.Power)
			continue
		} else if delegatee == nil {
			ctrler.logger.Debug("Not found validator", "address", types.Address(vote.Validator.Address), "power", vote.Validator.Power)
			continue
		}

		rwd := delegatee.DoReward(height, ctrlertypes.AmountPerPower(), ctrler.govParams.RewardPerPower())
		ctrler.logger.Debug("Block Reward", "address", delegatee.Addr, "reward", rwd.Dec())

		xerr = ctrler.delegateeLedger.SetFinality(delegatee)
		if xerr != nil {
			ctrler.logger.Error("Fail to finalize delegatee", "address", delegatee.Addr)
		}
	}
	return nil
}

func (ctrler *StakeCtrler) unfreezingStakes(height int64, acctHandler ctrlertypes.IAccountHandler) xerrors.XError {
	return ctrler.frozenLedger.IterateReadAllFinalityItems(func(s0 *Stake) xerrors.XError {
		if s0.RefundHeight <= height {
			// un-freezing s0
			// return not only s0.ReceivedReward but also s0.Amount
			xerr := acctHandler.Reward(
				s0.From,
				new(uint256.Int).Add(s0.Amount, s0.ReceivedReward),
				true)
			if xerr != nil {
				return xerr
			}

			_, _ = ctrler.frozenLedger.DelFinality(ledger.ToLedgerKey(s0.TxHash))
		}
		return nil
	})
}

func (ctrler *StakeCtrler) UpdateValidators(maxVals int) []abcitypes.ValidatorUpdate {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.updateValidators(maxVals)
}

// UpdateValidators is called after executing staking/unstaking txs and before committing the result of the txs executing.
// This means that the updated values of ledger is not committed yet.
// So, use ledger.IterateReadAllItems/IterateReadAllFinalityItems to get not changed values.
func (ctrler *StakeCtrler) updateValidators(maxVals int) []abcitypes.ValidatorUpdate {
	var allDelegatees DelegateeArray
	// NOTE:
	// IterateReadAllFinalityItems() returns delegatees, which are committed at previous block.
	// So, if staking tx is executed at block N,
	//     stake is saved(committed) at block N,
	//     it(updated validators) is notified to consensus engine at block N+1,
	//	   consensus add this account to validator set at block (N+1)+1+1.
	//	   (Refer to the comments in updateState(...) at github.com/tendermint/tendermint@v0.34.20/state/execution.go)
	// So, the account can sign a block from block N+3 in consensus engine
	if xerr := ctrler.delegateeLedger.IterateReadAllFinalityItems(func(d *Delegatee) xerrors.XError {
		// issue #59
		// Only delegatee who have deposited more than `MinValidatorStake` can become validator.
		if d.SelfAmount.Cmp(ctrler.govParams.MinValidatorStake()) >= 0 {
			allDelegatees = append(allDelegatees, d)
		}
		return nil
	}); xerr != nil {
		return nil
	}

	sort.Sort(PowerOrderDelegatees(allDelegatees)) // sort by power
	newValidators := selectValidators(allDelegatees, maxVals)

	sort.Sort(AddressOrderDelegatees(newValidators)) // sort by address
	upVals := validatorUpdates(ctrler.lastValidators, newValidators)

	// update lastValidators
	ctrler.lastValidators = newValidators
	return upVals
}

func validatorUpdates(existing, newers DelegateeArray) []abcitypes.ValidatorUpdate {
	valUpdates := make(abcitypes.ValidatorUpdates, 0, len(existing)+len(newers))

	i, j := 0, 0
	for i < len(existing) && j < len(newers) {
		ret := bytes.Compare(existing[i].Addr, newers[j].Addr)
		if ret < 0 {
			// this `existing` validator will be removed because it is not included in `newers`
			valUpdates = append(valUpdates, abcitypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
			i++
		} else if ret == 0 {
			if existing[i].TotalPower != newers[j].TotalPower {
				// if power is changed, add newer who has updated power
				valUpdates = append(valUpdates, abcitypes.UpdateValidator(newers[j].PubKey, int64(newers[j].TotalPower), "secp256k1"))
			} else {
				// if the power is not changed, exclude the validator in updated validators
			}
			i++
			j++
		} else { // ret > 0
			valUpdates = append(valUpdates, abcitypes.UpdateValidator(newers[j].PubKey, int64(newers[j].TotalPower), "secp256k1"))
			j++
		}
	}

	for ; i < len(existing); i++ {
		// removed
		valUpdates = append(valUpdates, abcitypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
	}
	for ; j < len(newers); j++ {
		// added newer
		valUpdates = append(valUpdates, abcitypes.UpdateValidator(newers[j].PubKey, int64(newers[j].TotalPower), "secp256k1"))
	}

	return valUpdates
}

func selectValidators(delegatees DelegateeArray, maxVals int) DelegateeArray {
	var validators DelegateeArray
	for i, val := range delegatees {
		if val.TotalPower <= 0 {
			break
		} else if i >= maxVals {
			break
		}
		validators = append(validators, val)
	}

	return validators
}

func (ctrler *StakeCtrler) Commit() ([]byte, int64, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if h0, v0, xerr := ctrler.delegateeLedger.Commit(); xerr != nil {
		return nil, -1, xerr
	} else if h1, v1, xerr := ctrler.frozenLedger.Commit(); xerr != nil {
		return nil, -1, xerr
	} else if v0 != v1 {
		return nil, -1, xerrors.ErrCommit.Wrapf("error: StakeCtrler.Commit() has wrong version number - v0:%v, v1:%v", v0, v1)
	} else {
		return crypto.DefaultHash(h0, h1), v0, nil
	}
}

func (ctrler *StakeCtrler) Close() xerrors.XError {
	if ctrler.delegateeLedger != nil {
		if xerr := ctrler.delegateeLedger.Close(); xerr != nil {
			ctrler.logger.Error("delegateeLedger.Close()", "error", xerr.Error())
		}
		ctrler.delegateeLedger = nil
	}
	if ctrler.frozenLedger != nil {
		if xerr := ctrler.frozenLedger.Close(); xerr != nil {
			ctrler.logger.Error("frozenLedger.Close()", "error", xerr.Error())
		}
		ctrler.frozenLedger = nil
	}
	return nil
}

// IStakeHandler's methods
func (ctrler *StakeCtrler) Validators() ([]*abcitypes.Validator, int64) {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	totalPower := int64(0)
	var ret []*abcitypes.Validator
	for _, v := range ctrler.lastValidators {
		totalPower += v.TotalPower
		ret = append(ret, &abcitypes.Validator{
			Address: v.Addr,
			Power:   int64(v.TotalPower),
		})
	}

	return ret, totalPower
}

func (ctrler *StakeCtrler) IsValidator(addr types.Address) bool {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	for _, v := range ctrler.lastValidators {
		if bytes.Compare(v.Addr, addr) == 0 {
			return true
		}
	}
	return false
}

func (ctrler *StakeCtrler) Delegatee(addr types.Address) *Delegatee {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(addr)); xerr != nil {
		return nil
	} else {
		return delegatee
	}
}

func (ctrler *StakeCtrler) PowerOf(addr types.Address) int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(addr)); xerr != nil {
		return 0
	} else if delegatee == nil {
		return 0
	} else {
		return delegatee.TotalPower
	}
}

func (ctrler *StakeCtrler) SelfPowerOf(addr types.Address) int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(addr)); xerr != nil {
		return 0
	} else if delegatee == nil {
		return 0
	} else {
		return delegatee.SelfPower
	}
}

func (ctrler *StakeCtrler) DelegatedPowerOf(addr types.Address) int64 {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(addr)); xerr != nil {
		return 0
	} else if delegatee == nil {
		return 0
	} else {
		return delegatee.TotalPower - delegatee.SelfPower
	}
}

func (ctrler *StakeCtrler) ReadTotalAmount() *uint256.Int {
	ret := uint256.NewInt(0)
	_ = ctrler.delegateeLedger.IterateReadAllFinalityItems(func(delegatee *Delegatee) xerrors.XError {
		_ = ret.Add(ret, delegatee.TotalAmount)
		return nil
	})
	return ret
}

func (ctrler *StakeCtrler) ReadTotalPower() int64 {
	ret := int64(0)
	_ = ctrler.delegateeLedger.IterateReadAllFinalityItems(func(delegatee *Delegatee) xerrors.XError {
		ret += delegatee.GetTotalPower()
		return nil
	})
	return ret
}

func (ctrler *StakeCtrler) ReadFrozenStakes() []*Stake {
	var ret []*Stake
	_ = ctrler.frozenLedger.IterateReadAllFinalityItems(func(s0 *Stake) xerrors.XError {
		ret = append(ret, s0)
		return nil
	})
	return ret
}

var _ ctrlertypes.ILedgerHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.ITrxHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.IBlockHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.IStakeHandler = (*StakeCtrler)(nil)
