package stake

import (
	"bytes"
	"github.com/cosmos/iavl"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	tmtypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
	"math/big"
	"sort"
	"sync"
)

type StakeCtrler struct {
	stakesDB      db.DB
	stakesTree    *iavl.MutableTree
	updatedStakes []*Stake // updated + newer

	delegateesDB      db.DB
	allDelegatees     DelegateeArray
	allDelegateesMap  map[account.AcctKey]*Delegatee // the key is delegatee's account key
	lastValidators    DelegateeArray
	updatedDelegatees DelegateeArray
	//removedDelegatees DelegateeArray

	frozenStakesDB  db.DB
	allFrozenStakes []*Stake
	newFrozenStakes []*Stake
	delFrozenStakes []*Stake

	logger log.Logger
	mtx    sync.RWMutex
}

func NewStakeCtrler(dbDir string, logger log.Logger) (*StakeCtrler, error) {

	// for all stakers
	stakeDB, err := db.NewDB("stakes", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}
	stakesTree, err := iavl.NewMutableTree(stakeDB, 1024)
	if err != nil {
		return nil, err
	}
	if _, err := stakesTree.Load(); err != nil {
		return nil, err
	}

	// load pubKey and fee reward
	delegateesDB, err := db.NewDB("delegatees", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}
	iterDelegateeDB, err := delegateesDB.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iterDelegateeDB.Close()

	var allDelegatees DelegateeArray
	allDelegateesMap := make(map[account.AcctKey]*Delegatee)

	for ; iterDelegateeDB.Valid(); iterDelegateeDB.Next() {
		k := iterDelegateeDB.Key()
		v := iterDelegateeDB.Value()
		f0 := NewFeeReward(nil)
		if err := f0.Decode(v); err != nil {
			return nil, err
		}

		acctKey := account.ToAcctKey(k)
		_, ok := allDelegateesMap[acctKey]
		if ok {
			logger.Error("the delegatee already exists")
			return nil, xerrors.New("delegatee is duplicated")
		}

		delegatee := NewDelegatee(k, f0.pubKey)
		// restore delegatee's public key and fee reward (#19)
		if err := delegatee.ReceivedReward.AddFeeReward(f0.receivedFee); err != nil {
			return nil, err
		}
		allDelegatees = append(allDelegatees, delegatee)
		allDelegateesMap[acctKey] = delegatee

	}
	if err := iterDelegateeDB.Error(); err != nil {
		return nil, err
	}

	// load stakes
	stopped, err := stakesTree.Iterate(func(key []byte, value []byte) bool {
		s0 := &Stake{}
		if err := json.Unmarshal(value, s0); err != nil {
			logger.Error("Unable to load stake", "txhash", types.HexBytes(key), "error", err)
			return true
		}

		if bytes.Compare(s0.TxHash, key) != 0 {
			logger.Error("Wrong TxHash", "key", types.HexBytes(key), "stake's txhash", s0.TxHash)
			return true
		}

		addrKey := account.ToAcctKey(s0.To)
		delegatee, ok := allDelegateesMap[addrKey]
		if !ok {
			logger.Error("not found delegatee")
			return true

			//delegatee = NewDelegatee(s0.To, nil)
			//allDelegatees = append(allDelegatees, delegatee)
			//allDelegateesMap[addrKey] = delegatee
		}

		// in AppendStake(), block reward is restored
		if err := delegatee.AppendStake(s0); err != nil {
			logger.Error("error in appending stake")
			return true
		}

		return false
	})
	if err != nil {
		return nil, xerrors.NewFrom(err)
	} else if stopped {
		return nil, xerrors.New("Stop to load stakers tree")
	}

	frozenDB, err := db.NewDB("frozen", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	iterFrozenDB, err := frozenDB.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iterFrozenDB.Close()

	var allFrozenStakes []*Stake
	for ; iterFrozenDB.Valid(); iterFrozenDB.Next() {
		v := iterFrozenDB.Value()
		s0 := &Stake{}
		if err := json.Unmarshal(v, s0); err != nil {
			return nil, err
		}
		allFrozenStakes = append(allFrozenStakes, s0)
	}
	if err := iterFrozenDB.Error(); err != nil {
		return nil, err
	}
	sort.Sort(refundHeightOrder(allFrozenStakes))

	ret := &StakeCtrler{
		stakesDB:         stakeDB,
		stakesTree:       stakesTree,
		delegateesDB:     delegateesDB,
		allDelegatees:    allDelegatees,
		allDelegateesMap: allDelegateesMap,
		frozenStakesDB:   frozenDB,
		allFrozenStakes:  allFrozenStakes,
		logger:           logger,
	}
	return ret, nil
}

func (ctrler *StakeCtrler) AddDelegatee(delegatee *Delegatee) *Delegatee {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.addDelegatee(delegatee)
}

func (ctrler *StakeCtrler) AddDelegateeWith(addr account.Address, pubKeyBytes types.HexBytes) *Delegatee {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	delegatee := NewDelegatee(addr, pubKeyBytes)
	return ctrler.addDelegatee(delegatee)
}

func (ctrler *StakeCtrler) addDelegatee(dlgtee *Delegatee) *Delegatee {
	addrKey := account.ToAcctKey(dlgtee.Addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; !ok {
		ctrler.allDelegateesMap[addrKey] = dlgtee
		ctrler.allDelegatees = append(ctrler.allDelegatees, dlgtee)
		return dlgtee
	}
	return nil
}

func (ctrler *StakeCtrler) removeDelegatee(addr account.Address) *Delegatee {
	addrKey := account.ToAcctKey(addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; ok {
		for i, dlgtee := range ctrler.allDelegatees {
			if bytes.Compare(dlgtee.Addr, addr) == 0 {
				delete(ctrler.allDelegateesMap, addrKey)
				ctrler.allDelegatees = append(ctrler.allDelegatees[:i], ctrler.allDelegatees[i+1:]...)
				return dlgtee
			}
		}

		ctrler.logger.Error("not same allDelegatees and allDelegateesMap)", "address", addr)
	}
	return nil
}

func (ctrler *StakeCtrler) GetDelegatee(idx int) *Delegatee {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if idx >= len(ctrler.allDelegatees) {
		return nil
	}

	return ctrler.allDelegatees[idx]
}

func (ctrler *StakeCtrler) DelegateeLen() int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return len(ctrler.allDelegateesMap)
}

func (ctrler *StakeCtrler) FindDelegatee(addr account.Address) *Delegatee {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findDelegatee(addr)
}

func (ctrler *StakeCtrler) findDelegatee(addr account.Address) *Delegatee {
	addrKey := account.ToAcctKey(addr)
	if delegatee, ok := ctrler.allDelegateesMap[addrKey]; ok {
		return delegatee
	}
	return nil
}

func (ctrler *StakeCtrler) validateStaking(ctx *trxs.TrxContext) error {
	if ctx.Tx.GetType() != trxs.TRX_STAKING {
		return xerrors.ErrInvalidTrxType
	}
	return nil
}

func (ctrler *StakeCtrler) validateUnstaking(ctx *trxs.TrxContext) error {
	if ctx.Tx.GetType() != trxs.TRX_UNSTAKING {
		return xerrors.ErrInvalidTrxType
	}
	return nil
}

func (ctrler *StakeCtrler) Validate(ctx *trxs.TrxContext) error {
	if ctx.Tx.GetType() != trxs.TRX_STAKING && ctx.Tx.GetType() != trxs.TRX_STAKING {
		return xerrors.ErrInvalidTrxType
	} else {
		return nil //return ctrler.validateUnstaking(ctx)
	}
}

func (ctrler *StakeCtrler) DelegateStake(to *Delegatee, s0 *Stake) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.delegateStake(to, s0)
}

func (ctrler *StakeCtrler) delegateStake(to *Delegatee, s0 *Stake) {
	if xerr := to.AppendStake(s0); xerr != nil {
		// Not reachable. AppendStake() does not return error
		ctrler.logger.Error("Not reachable", "error", xerr)
		panic(xerr)
	}
	ctrler.updatedStakes = append(ctrler.updatedStakes, s0)
}

func (ctrler *StakeCtrler) GetFrozenStakes() []*Stake {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.allFrozenStakes
}

func (ctrler *StakeCtrler) ProcessFrozenStakesAt(height int64, acctFinder account.IAccountFinder) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// ctrler.allFrozenStakes is ordered by RefundHeight
	for _, s0 := range ctrler.allFrozenStakes {
		if height >= s0.RefundHeight {
			if acct := acctFinder.FindAccount(s0.From, true); acct == nil {
				return xerrors.ErrNotFoundAccount
			} else if xerr := acct.AddBalance(new(big.Int).Add(s0.Amount, s0.ReceivedReward)); xerr != nil {
				return xerr
			}
			ctrler.delFrozenStakes = append(ctrler.delFrozenStakes, s0)
		} else {
			break
		}
	}
	return nil
}

func (ctrler *StakeCtrler) GetTotalAmount() *big.Int {
	// todo: improve performance
	amt := big.NewInt(0)
	for _, s0 := range ctrler.allDelegatees {
		amt = new(big.Int).Add(amt, s0.GetTotalAmount())
	}
	return amt
}

func (ctrler *StakeCtrler) GetTotalPower() int64 {
	// todo: improve performance
	power := int64(0)
	for _, s0 := range ctrler.allDelegatees {
		power += s0.GetTotalPower()
	}
	return power
}

func (ctrler *StakeCtrler) GetTotalPowerOf(addr account.Address) int64 {
	power := int64(0)
	if delegatee, ok := ctrler.allDelegateesMap[account.ToAcctKey(addr)]; ok {
		power += delegatee.GetTotalPower()
	}
	return power
}

func selectValidators(delegatees DelegateeArray, maxVals int) DelegateeArray {
	sort.Sort(powerOrderedDelegatees(delegatees)) // sort by power

	n := libs.MIN(len(delegatees), maxVals)
	validators := make(DelegateeArray, n)
	copy(validators, delegatees[:n])

	return validators
}

func (ctrler *StakeCtrler) UpdateValidators(maxVals int) []tmtypes.ValidatorUpdate {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	newValidators := selectValidators(ctrler.allDelegatees, maxVals)
	sort.Sort(addressOrderedDelegatees(newValidators)) // sort by address

	ctrler.lastValidators = newValidators
	return validatorUpdates(ctrler.lastValidators, newValidators)
}

func (ctrler *StakeCtrler) GetLastValidatorCnt() int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return len(ctrler.lastValidators)
}

func (ctrler *StakeCtrler) IsValidator(addr account.Address) bool {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	for _, val := range ctrler.lastValidators {
		if bytes.Compare(val.Addr, addr) == 0 {
			return true
		}
	}
	return false
}

func validatorUpdates(existing, newers DelegateeArray) []tmtypes.ValidatorUpdate {
	valUpdates := make(tmtypes.ValidatorUpdates, 0, len(existing)+len(newers))

	i, j := 0, 0
	for i < len(existing) && j < len(newers) {
		ret := bytes.Compare(existing[i].Addr, newers[j].Addr)
		if ret < 0 {
			// lastValidators[i] is removed
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
			i++
		} else if ret == 0 {
			if existing[i].TotalPower != newers[j].TotalPower {
				// lastValidator[i] is updated to newValidators[j]
				valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].TotalPower, "secp256k1"))
			}
			i++
			j++
		} else { // ret > 0
			// newValidators[j] is added
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].TotalPower, "secp256k1"))
			j++
		}
	}

	for ; i < len(existing); i++ {
		// removed
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
	}
	for ; j < len(newers); j++ {
		// added newer
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].TotalPower, "secp256k1"))
	}

	return valUpdates
}

func (ctrler *StakeCtrler) execStaking(ctx *trxs.TrxContext) error {
	if err := ctx.Sender.SubBalance(ctx.Tx.Amount); err != nil {
		return err
	}

	if ctx.Exec {
		delegatee := ctrler.findDelegatee(ctx.Tx.To)
		if delegatee == nil && bytes.Compare(ctx.Tx.From, ctx.Tx.To) == 0 {
			// new delegatee (staking to my self)
			delegatee = ctrler.addDelegatee(NewDelegatee(ctx.Tx.From, ctx.SenderPubKey))
		} else if delegatee == nil {
			// there is no delegatee whose address is ctx.Tx.To
			return xerrors.ErrNotFoundStaker
		}

		s0 := NewStakeWithAmount(ctx.Tx.From, ctx.Tx.To, ctx.Tx.Amount, ctx.Height, ctx.TxHash, ctx.GovRuleHandler)
		expectedPower := ctrler.GetTotalPower() + s0.Power
		if expectedPower < 0 || expectedPower > ctx.GovRuleHandler.MaxTotalPower() {
			return xerrors.ErrTooManyPower
		}

		ctrler.delegateStake(delegatee, s0)
		ctrler.updatedDelegatees = append(ctrler.updatedDelegatees, delegatee)
	}

	return nil
}

func (ctrler *StakeCtrler) execUnstaking(ctx *trxs.TrxContext) error {
	if ctx.Exec {
		delegatee := ctrler.findDelegatee(ctx.Tx.To)
		if delegatee == nil {
			return xerrors.ErrNotFoundStaker
		}

		txhash := ctx.Tx.Payload.(*trxs.TrxPayloadUnstaking).TxHash
		if txhash == nil && len(txhash) != 32 {
			return xerrors.ErrInvalidTrxPayloadParams
		}
		s0 := delegatee.DelStake(txhash)
		if s0 == nil {
			return xerrors.ErrNotFoundStake
		}

		s0.RefundHeight = ctx.Height + ctx.GovRuleHandler.GetLazyRewardBlocks()
		// ctrler.frozenStakes is ordered by RefundHeight
		ctrler.newFrozenStakes = append(ctrler.newFrozenStakes, s0)

		if delegatee.SelfPower == 0 {
			// issue #12
			stakes := delegatee.DelAllStakes()
			ctrler.newFrozenStakes = append(ctrler.newFrozenStakes, stakes...)
		}

		if delegatee.TotalPower == 0 {
			// issue #11 : add and freeze a stake for FEE

			feeAmt := delegatee.ReceivedReward.FeeReward()
			feeStake := &Stake{
				From:           delegatee.Addr,
				To:             delegatee.Addr,
				ReceivedReward: feeAmt,
				RefundHeight:   ctx.Height + ctx.GovRuleHandler.GetLazyRewardBlocks(),
			}
			ctrler.newFrozenStakes = append(ctrler.newFrozenStakes, feeStake)

			// issue #19
			// it will be updated(saved) in Commit()
			if err := delegatee.ReceivedReward.SubFeeReward(feeAmt); err != nil {
				return err
			}

			_ = ctrler.removeDelegatee(delegatee.Addr)
		}
		ctrler.updatedDelegatees = append(ctrler.updatedDelegatees, delegatee)
	}
	return nil
}

func (ctrler *StakeCtrler) ApplyReward(feeOwner account.Address, fee *big.Int) *big.Int {

	// NOTE!!!
	// Because ApplyReward() is called after executing un-staking-tx,
	// the account who submit un-staking-tx at this block CAN NOT be rewarded for the previous block.
	// Likewise, because ApplyReward() is called after executing staking-tx,
	// the account who submit staking-tx (delegates to current validator) at this block is rewarded for previous block.

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	blockReward := big.NewInt(0)
	for _, val := range ctrler.lastValidators {
		// block reward : fee is not included
		blockReward = blockReward.Add(blockReward, val.ApplyBlockReward())
		ctrler.updatedStakes = append(ctrler.updatedStakes, val.GetAllStakes()...)

		// fee reward
		if bytes.Compare(val.Addr, feeOwner) == 0 {
			val.ApplyFeeReward(fee)

			// update delegatee's fee reward (#19)
			ctrler.updatedDelegatees = append(ctrler.updatedDelegatees, val)
		}
	}
	return blockReward
}

func (ctrler *StakeCtrler) Execute(ctx *trxs.TrxContext) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch ctx.Tx.GetType() {
	case trxs.TRX_STAKING:
		return ctrler.execStaking(ctx)
	case trxs.TRX_UNSTAKING:
		return ctrler.execUnstaking(ctx)
	default:
		return xerrors.New("unknown transaction type")
	}
}

func (ctrler *StakeCtrler) Commit() ([]byte, int64, error) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// todo: Commit() should be atomic operation

	frozenBatch := ctrler.frozenStakesDB.NewBatch()
	defer frozenBatch.Close()

	for _, s := range ctrler.delFrozenStakes {
		if err := frozenBatch.Delete(s.TxHash); err != nil {
			return nil, -1, err
		}
	}
	ctrler.allFrozenStakes = ctrler.allFrozenStakes[len(ctrler.delFrozenStakes):]
	ctrler.delFrozenStakes = nil
	for _, s := range ctrler.newFrozenStakes {
		if bz, err := json.Marshal(s); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		} else if err := frozenBatch.Set(s.TxHash, bz); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		} else if _, _, err := ctrler.stakesTree.Remove(s.TxHash); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		}
	}
	ctrler.allFrozenStakes = append(ctrler.allFrozenStakes, ctrler.newFrozenStakes...)
	ctrler.newFrozenStakes = nil
	sort.Sort(refundHeightOrder(ctrler.allFrozenStakes))

	for _, s0 := range ctrler.updatedStakes {
		if bz, err := json.Marshal(s0); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		} else if _, err := ctrler.stakesTree.Set(s0.TxHash, bz); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		}
	}
	ctrler.updatedStakes = nil

	// updatedDelegatees has delegatees who are newly added/removed or rewarded
	for _, d0 := range ctrler.updatedDelegatees {
		if d0.TotalPower == 0 {
			// remove delegatee
			if err := ctrler.delegateesDB.Delete(d0.Addr); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			}
		} else {
			if bz, err := d0.ReceivedReward.fee.Encode(); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			} else if err := ctrler.delegateesDB.Set(d0.Addr, bz); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			}
		}
	}
	ctrler.updatedDelegatees = nil

	if err := frozenBatch.Write(); err != nil {
		return nil, -1, err
	}

	return ctrler.stakesTree.SaveVersion()
}

func (ctrler *StakeCtrler) Close() error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.stakesDB != nil {
		if err := ctrler.stakesDB.Close(); err != nil {
			return nil
		}
	}
	if ctrler.frozenStakesDB != nil {
		if err := ctrler.frozenStakesDB.Close(); err != nil {
			return err
		}
	}

	ctrler.stakesDB = nil
	ctrler.stakesTree = nil
	ctrler.frozenStakesDB = nil
	return nil
}

var _ trxs.ITrxHandler = (*StakeCtrler)(nil)
var _ types.ILedgerHandler = (*StakeCtrler)(nil)
var _ types.IStakeHandler = (*StakeCtrler)(nil)

type DelegateeArray []*Delegatee

func (vs DelegateeArray) SumTotalAmount() *big.Int {
	var amt *big.Int
	for _, val := range vs {
		amt = new(big.Int).Add(amt, val.TotalAmount)
	}
	return amt
}

func (vs DelegateeArray) SumTotalPower() int64 {
	power := int64(0)
	for _, val := range vs {
		power += val.TotalPower
	}
	return power
}

func (vs DelegateeArray) SumTotalReward() *big.Int {
	var reward *big.Int
	for _, val := range vs {
		reward = new(big.Int).Add(reward, val.GetTotalReward())
	}
	return reward
}

func (vs DelegateeArray) SumBlockReward() *big.Int {
	var reward *big.Int
	for _, val := range vs {
		reward = new(big.Int).Add(reward, val.ReceivedReward.BlockReward())
	}
	return reward
}

func (vs DelegateeArray) SumTotalFeeReward() *big.Int {
	var fee *big.Int
	for _, val := range vs {
		fee = new(big.Int).Add(fee, val.ReceivedReward.FeeReward())
	}
	return fee
}

type powerOrderedDelegatees []*Delegatee

func (vs powerOrderedDelegatees) Len() int {
	return len(vs)
}

// descending order by TotalPower
func (vs powerOrderedDelegatees) Less(i, j int) bool {
	if vs[i].TotalPower != vs[j].TotalPower {
		return vs[i].TotalPower > vs[j].TotalPower
	} else if len(vs[i].Stakes) != len(vs[j].Stakes) {
		return len(vs[i].Stakes) > len(vs[j].Stakes)
	} else if bytes.Compare(vs[i].Addr, vs[j].Addr) > 0 {
		return true
	}
	return false
}

func (vs powerOrderedDelegatees) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (powerOrderedDelegatees)(nil)

type addressOrderedDelegatees []*Delegatee

func (vs addressOrderedDelegatees) Len() int {
	return len(vs)
}

// ascending order by address
func (vs addressOrderedDelegatees) Less(i, j int) bool {

	return bytes.Compare(vs[i].Addr, vs[j].Addr) < 0
}

func (vs addressOrderedDelegatees) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (addressOrderedDelegatees)(nil)
