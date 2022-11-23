package stake

import (
	"bytes"
	"github.com/cosmos/iavl"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
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

type ValidatorInfo struct {
	Address types.Address
	PubKey  types.HexBytes
	Power   int64
}

type ValidatorInfoList []*ValidatorInfo

func (vilst ValidatorInfoList) Len() int {
	return len(vilst)
}

// ascending order by Address
func (vilst ValidatorInfoList) Less(i, j int) bool {
	return bytes.Compare(vilst[i].Address, vilst[j].Address) < 0
}

func (vilst ValidatorInfoList) Swap(i, j int) {
	vilst[i], vilst[j] = vilst[j], vilst[i]
}

var _ sort.Interface = (ValidatorInfoList)(nil)

func (vilst ValidatorInfoList) isExist(addr types.Address) bool {
	for _, vi := range vilst {
		if bytes.Compare(vi.Address, addr) == 0 {
			return true
		}
	}
	return false
}

func (vilst ValidatorInfoList) find(addr types.Address) *ValidatorInfo {
	for _, vi := range vilst {
		if bytes.Compare(vi.Address, addr) == 0 {
			return vi
		}
	}
	return nil
}

type StakeCtrler struct {
	stakesDB      db.DB
	stakesTree    *iavl.MutableTree
	updatedStakes []*Stake // updated + newer

	delegateesDB      db.DB
	delegateesDBBatch db.Batch
	allDelegatees     DelegateeArray
	allDelegateesMap  map[types.AcctKey]*Delegatee // the key is delegatee's account key

	lastValidators ValidatorInfoList

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

	// for all delegatees
	delegateesDB, err := db.NewDB("delegatees", "govleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	// load delegatees
	iterDelegatees, err := delegateesDB.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iterDelegatees.Close()

	var allDelegatees DelegateeArray
	allDelegateesMap := make(map[types.AcctKey]*Delegatee)

	for ; iterDelegatees.Valid(); iterDelegatees.Next() {
		k := iterDelegatees.Key()
		if _, ok := allDelegateesMap[types.ToAcctKey(k)]; ok {
			return nil, xerrors.New("duplicated delegatee")
		}
		v := iterDelegatees.Value()
		dgtee := &Delegatee{}
		if err := json.Unmarshal(v, dgtee); err != nil {
			return nil, err
		}
		allDelegatees = append(allDelegatees, dgtee)
		allDelegateesMap[types.ToAcctKey(k)] = dgtee

	}
	if err := iterDelegatees.Error(); err != nil {
		return nil, err
	}
	sort.Sort(powerOrderedDelegatees(allDelegatees))

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

		addrKey := types.ToAcctKey(s0.To)
		delegatee, ok := allDelegateesMap[addrKey]
		if !ok {
			logger.Error("Not found delegatee", "address", types.HexBytes(s0.To))
			return true
		}

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
		stakesDB:          stakeDB,
		stakesTree:        stakesTree,
		delegateesDB:      delegateesDB,
		delegateesDBBatch: delegateesDB.NewBatch(),
		allDelegatees:     allDelegatees,
		allDelegateesMap:  allDelegateesMap,
		frozenStakesDB:    frozenDB,
		allFrozenStakes:   allFrozenStakes,
		logger:            logger,
	}
	return ret, nil
}

func (ctrler *StakeCtrler) AddDelegatee(delegatee *Delegatee) *Delegatee {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.addDelegatee(delegatee)
}

func (ctrler *StakeCtrler) AddDelegateeWith(addr types.Address, pubKeyBytes types.HexBytes) *Delegatee {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	delegatee := NewDelegatee(addr, pubKeyBytes)
	return ctrler.addDelegatee(delegatee)
}

func (ctrler *StakeCtrler) addDelegatee(delegatee *Delegatee) *Delegatee {
	addrKey := types.ToAcctKey(delegatee.Addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; !ok {
		if blob, err := json.Marshal(delegatee); err != nil {
			panic(err)
		} else if err := ctrler.delegateesDBBatch.Set(addrKey[:], blob); err != nil {
			panic(err)
		}
		ctrler.allDelegatees = append(ctrler.allDelegatees, delegatee)
		ctrler.allDelegateesMap[addrKey] = delegatee
		return delegatee
	}
	return nil
}

func (ctrler *StakeCtrler) removeDelegatee(addr types.Address) *Delegatee {
	addrKey := types.ToAcctKey(addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; ok {
		for i, staker := range ctrler.allDelegatees {
			if bytes.Compare(staker.Addr, addr) == 0 {
				if err := ctrler.delegateesDBBatch.Delete(addrKey[:]); err != nil {
					panic(err)
				}
				ctrler.allDelegatees = append(ctrler.allDelegatees[:i], ctrler.allDelegatees[i+1:]...)
				delete(ctrler.allDelegateesMap, addrKey)
				return staker
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

	return len(ctrler.allDelegatees)
}

func (ctrler *StakeCtrler) FindDelegatee(addr types.Address) *Delegatee {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findDelegatee(addr)
}

func (ctrler *StakeCtrler) findDelegatee(addr types.Address) *Delegatee {
	addrKey := types.ToAcctKey(addr)
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

func (ctrler *StakeCtrler) applyStaking(ctx *trxs.TrxContext) error {
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

		//if xerr := delegatee.AppendStake(s0); xerr != nil {
		//	// Not reachable. AppendStake() does not return error
		//	ctrler.logger.Error("Not reachable", "error", xerr)
		//	panic(xerr)
		//}
		//
		//ctrler.updatedStakes = append(ctrler.updatedStakes, s0)
		ctrler.delegateStake(delegatee, s0)
	}

	return nil
}

func (ctrler *StakeCtrler) applyUnstakingByTxHash(ctx *trxs.TrxContext) error {
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
			_ = ctrler.removeDelegatee(delegatee.Addr)

			// issue #11
			feeStake := &Stake{
				From:         delegatee.Addr,
				To:           delegatee.Addr,
				Reward:       delegatee.FeeReward,
				RefundHeight: 0,
			}
			ctrler.newFrozenStakes = append(ctrler.newFrozenStakes, feeStake)
		}
	}
	return nil
}

// applyUnstaking() is DEPRECATED.
func (ctrler *StakeCtrler) applyUnstaking(ctx *trxs.TrxContext) error {
	if ctx.Exec {
		staker := ctrler.findDelegatee(ctx.Tx.To)
		if staker == nil {
			return xerrors.ErrNotFoundStaker
		}

		damt := ctx.Tx.Amount
		if damt.Cmp(staker.TotalAmount) > 0 {
			return xerrors.New("invalid unstaking amount")
		}

		stakes := staker.StakesOf(ctx.Tx.From)
		if stakes == nil {
			return xerrors.ErrNotFoundStake
		}

		for _, s0 := range stakes {
			if damt.Cmp(s0.Amount) >= 0 {
				// remove it
				_ = staker.DelStake(s0.TxHash) // returns `*Stake` same as `s0`
				damt = new(big.Int).Sub(damt, s0.Amount)

				s0.RefundHeight = ctx.Height + ctx.GovRuleHandler.GetLazyRewardBlocks()

				// ctrler.frozenStakes is ordered by RefundHeight
				ctrler.newFrozenStakes = append(ctrler.newFrozenStakes, s0)
			} else {
				//
				// Now, partial un-staking is not supported.
				// todo: implement partial unstaking

				//s.DecreaseAmount(damt)
				ctrler.logger.Debug("Not supported partial unstaking")

				break
			}
		}

		// staker.TotalPower, staker.SumTotalAmount, staker.Stakes.Len() is related...
		// todo: these variables (TotalPower, TotalAmount, Stakes length) should be checked.
		if staker.TotalPower == 0 {
			_ = ctrler.removeDelegatee(staker.Addr)
		}
	}
	return nil
}

func (ctrler *StakeCtrler) Execute(ctx *trxs.TrxContext) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch ctx.Tx.GetType() {
	case trxs.TRX_STAKING:
		return ctrler.applyStaking(ctx)
	case trxs.TRX_UNSTAKING:
		return ctrler.applyUnstakingByTxHash(ctx)
	default:
		return xerrors.New("unknown transaction type")
	}
}

func (ctrler *StakeCtrler) ApplyReward(feeOwner types.Address, fee *big.Int) *big.Int {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	reward := big.NewInt(0)
	for _, dlgtee := range ctrler.allDelegatees {
		if bytes.Compare(dlgtee.Addr, feeOwner) == 0 && fee.Sign() > 0 {
			dlgtee.ApplyFeeReward(fee)
		}

		// reward does not include fee
		reward = reward.Add(reward, dlgtee.ApplyReward())
	}
	return reward
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

func (ctrler *StakeCtrler) ProcessFrozenStakesAt(height int64, acctFinder types.IAccountFinder) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// ctrler.allFrozenStakes is ordered by RefundHeight
	for _, s0 := range ctrler.allFrozenStakes {
		if height >= s0.RefundHeight {
			if acct := acctFinder.FindAccount(s0.From, true); acct == nil {
				return xerrors.ErrNotFoundAccount
			} else if xerr := acct.AddBalance(new(big.Int).Add(s0.Amount, s0.Reward)); xerr != nil {
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

func (ctrler *StakeCtrler) GetTotalPowerOf(addr types.Address) int64 {
	power := int64(0)
	if delegatee, ok := ctrler.allDelegateesMap[types.ToAcctKey(addr)]; ok {
		power += delegatee.GetTotalPower()
	}
	return power
}

func (ctrler *StakeCtrler) UpdateValidators(maxVals int) []tmtypes.ValidatorUpdate {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	sort.Sort(powerOrderedDelegatees(ctrler.allDelegatees)) // sort by power

	var newValidators ValidatorInfoList
	n := libs.MIN(len(ctrler.allDelegatees), maxVals)
	for i := 0; i < n; i++ {
		delegatee := ctrler.allDelegatees[i]
		newValidators = append(newValidators, &ValidatorInfo{
			Address: delegatee.Addr,
			PubKey:  delegatee.PubKey,
			Power:   delegatee.TotalPower,
		})
	}
	sort.Sort(newValidators) // sort by address

	valUps := validatorUpdates(ctrler.lastValidators, newValidators)
	ctrler.lastValidators = newValidators

	return valUps
}

func (ctrler *StakeCtrler) GetLastValidatorCnt() int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return len(ctrler.lastValidators)
}

func (ctrler *StakeCtrler) IsValidator(addr types.Address) bool {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	for _, val := range ctrler.lastValidators {
		if bytes.Compare(val.Address, addr) == 0 {
			return true
		}
	}
	return false
}

func validatorUpdates(existing, newers ValidatorInfoList) []tmtypes.ValidatorUpdate {
	valUpdates := make(tmtypes.ValidatorUpdates, 0, len(existing)+len(newers))

	i, j := 0, 0
	for i < len(existing) && j < len(newers) {
		ret := bytes.Compare(existing[i].Address, newers[j].Address)
		if ret < 0 {
			// lastValidators[i] is removed
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
			i++
		} else if ret == 0 {
			if existing[i].Power != newers[j].Power {
				// lastValidator[i] is updated to newValidators[j]
				valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
			}
			i++
			j++
		} else { // ret > 0
			// newValidators[j] is added
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
			j++
		}
	}

	for ; i < len(existing); i++ {
		// removed
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
	}
	for ; j < len(newers); j++ {
		// added newer
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
	}

	return valUpdates
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

	if err := frozenBatch.Write(); err != nil {
		return nil, -1, err
	}

	if err := ctrler.delegateesDBBatch.Write(); err != nil {
		return nil, -1, err
	} else if err := ctrler.delegateesDBBatch.Close(); err != nil {
		panic(err)
	}
	ctrler.delegateesDBBatch = ctrler.delegateesDB.NewBatch()

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
		reward = new(big.Int).Add(reward, val.TotalReward)
	}
	return reward
}

func (vs DelegateeArray) SumTotalFeeReward() *big.Int {
	var fee *big.Int
	for _, val := range vs {
		fee = new(big.Int).Add(fee, val.FeeReward)
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
