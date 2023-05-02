package stake

import (
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
	"sync"
)

type StakeCtrler struct {
	lastValidators  DelegateeArray
	delegateeLedger ledger.IFinalityLedger[*Delegatee]
	frozenLedger    ledger.IFinalityLedger[*Stake]

	govHelper ctrlertypes.IGovHelper
	logger    tmlog.Logger
	mtx       sync.RWMutex
}

func NewStakeCtrler(config *cfg.Config, govHelper ctrlertypes.IGovHelper, logger tmlog.Logger) (*StakeCtrler, error) {
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
		govHelper:    govHelper,
		logger:       logger,
	}
	return ret, nil
}

func (ctrler *StakeCtrler) InitLedger(req interface{}) xerrors.XError {
	// init validators
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	validators, ok := req.([]abcitypes.ValidatorUpdate)
	if !ok {
		return xerrors.ErrInitChain.Wrapf("wrong parameter: StakeCtrler::InitLedger() requires []ValidatorUpdates")
	}

	for _, val := range validators {
		pubBytes := val.PubKey.GetSecp256K1()
		addr, xerr := crypto.PubBytes2Addr(pubBytes)
		if xerr != nil {
			return xerr
		}
		s0 := NewStakeWithPower(
			addr, addr, // self staking
			val.Power,
			0,
			bytes.ZeroBytes(32), // 0x00... txhash
			ctrler.govHelper)

		d := NewDelegatee(addr, pubBytes)
		if xerr := d.AddStake(s0); xerr != nil {
			return xerr
		} else if xerr := ctrler.delegateeLedger.SetFinality(d); xerr != nil {
			return xerr
		}

		// To give the first block reward to genesis validators,
		// run the following code
		// ctrler.lastValidators = append(ctrler.lastValidators, d)
	}

	return nil
}

func (ctrler *StakeCtrler) ValidateTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_STAKING:
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

	return nil
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
		// self staking
		// add new delegatee
		delegatee = NewDelegatee(ctx.Tx.From, ctx.SenderPubKey)
	} else if delegatee == nil {
		// there is no delegatee whose address is ctx.Tx.To
		return xerrors.ErrNotFoundDelegatee
	}

	// create stake and delegate it to delegatee
	s0 := NewStakeWithAmount(ctx.Tx.From, ctx.Tx.To, ctx.Tx.Amount, ctx.Height, ctx.TxHash, ctx.GovHelper)
	if xerr := delegatee.addStake(s0); xerr != nil {
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
	if txhash == nil && len(txhash) != 32 {
		return xerrors.ErrInvalidTrxPayloadParams
	}
	s0 := delegatee.DelStake(txhash)
	if s0 == nil {
		return xerrors.ErrNotFoundStake
	}

	s0.RefundHeight = ctx.Height + ctx.GovHelper.LazyRewardBlocks()
	_ = setUpdateFrozen(s0) // add s0 to frozen ledger

	if delegatee.SelfPower == 0 {
		stakes := delegatee.DelAllStakes()
		for _, s0 := range stakes {
			s0.RefundHeight = ctx.Height + ctx.GovHelper.LazyRewardBlocks()
			_ = setUpdateFrozen(s0) // add s0 to frozen ledger
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

func (ctrler *StakeCtrler) ValidateBlock(ctx *ctrlertypes.BlockContext) xerrors.XError {
	// do nothing
	return nil
}

func (ctrler *StakeCtrler) ExecuteBlock(ctx *ctrlertypes.BlockContext) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.logger.Debug("StakeCtrler-ExecuteBlock", "height", ctx.BlockInfo.Header.Height)

	if ctx.Fee.Sign() > 0 {
		if xerr := ctrler.doReward(); xerr != nil {
			return xerr
		}
	}
	if xerr := ctrler.unfreezingStakes(ctx.BlockInfo.Header.Height, ctx.AcctHelper); xerr != nil {
		return xerr
	}
	ctx.ValUpdates = ctrler.updateValidators(int(ctx.GovHelper.MaxValidatorCnt()))

	return nil
}

// the following functions are called in ExecuteBlock()
//

func (ctrler *StakeCtrler) doReward() xerrors.XError {

	// todo: implement completely: issue #29
	// GetFinality() returns delegatee which is updated in execStaking / execUnstaking.
	// So, the stakes, that is staked or un-staked at this block, is rewarded. is it right?
	for _, val := range ctrler.lastValidators {
		if delegatee, xerr := ctrler.delegateeLedger.GetFinality(ledger.ToLedgerKey(val.Addr)); xerr != nil {
			return xerr
		} else if delegatee == nil {
			return xerrors.ErrNotFoundDelegatee
		} else {
			_ = delegatee.DoReward()
			_ = ctrler.delegateeLedger.SetFinality(delegatee)
		}
	}
	return nil
}

func (ctrler *StakeCtrler) unfreezingStakes(height int64, acctHelper ctrlertypes.IAccountHelper) xerrors.XError {
	return ctrler.frozenLedger.IterateAllFinalityItems(func(s0 *Stake) xerrors.XError {
		if s0.RefundHeight <= height {
			// un-freezing s0
			if xerr := acctHelper.Reward(s0.From, s0.ReceivedReward, true); xerr != nil {
				return xerr
			}

			_, _ = ctrler.frozenLedger.DelFinality(ledger.ToLedgerKey(s0.TxHash))
		}
		return nil
	})
}

// UpdateValidators is called after executing staking/unstaking txs, before committing the result of the txs executing
// This means that the updated values of ledger is not committed yet.
// So ledger.IterateAllItems/IterateAllFinalityItems returns not updated values.
func (ctrler *StakeCtrler) updateValidators(maxVals int) []abcitypes.ValidatorUpdate {
	var allDelegatees DelegateeArray
	// NOTE:
	// IterateAllFinalityItems() returns delegatees, which are committed at previous block.
	// So, if staking tx is executed at block N,
	//     stake is saved(committed) at block N,
	//     it(updated validators) is notified to consensus engine at block N+1,
	//	   consensus add this account to validator set at block (N+1)+1+1.
	//	   (Refer to the comments in updateState(...) at github.com/tendermint/tendermint@v0.34.20/state/execution.go)
	// So, the account can sign a block from block N+3 in consensus engine
	if xerr := ctrler.delegateeLedger.IterateAllFinalityItems(func(d *Delegatee) xerrors.XError {
		allDelegatees = append(allDelegatees, d)
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
			// this 'existing' validator will be removed because the power is 0
			valUpdates = append(valUpdates, abcitypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
			i++
		} else if ret == 0 {
			if existing[i].TotalPower != newers[j].TotalPower {
				// if power is changed, add newser
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

	ctrler.logger.Debug("StakeCtrler-Commit")

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
			ctrler.logger.Error("StakeCtrler", "delegateeLedger.Close() returns error", xerr.Error())
		}
		ctrler.delegateeLedger = nil
	}
	if ctrler.frozenLedger != nil {
		if xerr := ctrler.frozenLedger.Close(); xerr != nil {
			ctrler.logger.Error("StakeCtrler", "frozenLedger.Close() returns error", xerr.Error())
		}
		ctrler.frozenLedger = nil
	}
	return nil
}

// IStakeHelper's methods
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
	for _, v := range ctrler.lastValidators {
		if bytes.Compare(v.Addr, addr) == 0 {
			return true
		}
	}
	return false
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

func (ctrler *StakeCtrler) GetTotalAmount() *uint256.Int {
	ret := uint256.NewInt(0)
	_ = ctrler.delegateeLedger.IterateAllFinalityItems(func(delegatee *Delegatee) xerrors.XError {
		_ = ret.Add(ret, delegatee.TotalAmount)
		return nil
	})
	return ret
}

func (ctrler *StakeCtrler) GetTotalPower() int64 {
	ret := int64(0)
	_ = ctrler.delegateeLedger.IterateAllFinalityItems(func(delegatee *Delegatee) xerrors.XError {
		ret += delegatee.GetTotalPower()
		return nil
	})
	return ret
}

func (ctrler *StakeCtrler) GetFrozenStakes() []*Stake {
	var ret []*Stake
	_ = ctrler.frozenLedger.IterateAllFinalityItems(func(s0 *Stake) xerrors.XError {
		ret = append(ret, s0)
		return nil
	})
	return ret
}

var _ ctrlertypes.ILedgerHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.ITrxHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.IBlockHandler = (*StakeCtrler)(nil)
var _ ctrlertypes.IStakeHelper = (*StakeCtrler)(nil)
