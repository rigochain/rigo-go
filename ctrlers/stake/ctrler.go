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

	allDelegatees    DelegateeArray
	allDelegateesMap map[types.AcctKey]*Delegatee // the key is staker's account key

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

	var allStakers DelegateeArray
	allStakersMap := make(map[types.AcctKey]*Delegatee)

	stopped, err := stakesTree.Iterate(func(key []byte, value []byte) bool {
		s0 := &Stake{}
		if err := json.Unmarshal(value, s0); err != nil {
			logger.Error("Unable to load staker", "address", types.HexBytes(key), "error", err)
			return true
		}

		if bytes.Compare(s0.TxHash, key) != 0 {
			logger.Error("Wrong TxHash", "key", types.HexBytes(key), "stake's txhash", s0.TxHash)
			return true
		}

		addrKey := types.ToAcctKey(s0.To)
		delegatee, ok := allStakersMap[addrKey]
		if !ok {
			delegatee = &Delegatee{
				Addr: s0.To,
			}
			allStakers = append(allStakers, delegatee)
			allStakersMap[addrKey] = delegatee
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

	sort.Sort(powerOrderedStakeSets(allStakers))

	frozenDB, err := db.NewDB("frozen", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	iter, err := frozenDB.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var allFrozenStakes []*Stake
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		s0 := &Stake{}
		if err := json.Unmarshal(v, s0); err != nil {
			return nil, err
		}
		allFrozenStakes = append(allFrozenStakes, s0)
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	sort.Sort(refundHeightOrder(allFrozenStakes))

	ret := &StakeCtrler{
		stakesDB:         stakeDB,
		stakesTree:       stakesTree,
		allDelegatees:    allStakers,
		allDelegateesMap: allStakersMap,
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

func (ctrler *StakeCtrler) AddDelegateeWith(addr types.Address, pubKeyBytes types.HexBytes) *Delegatee {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	staker := NewDelegatee(addr, pubKeyBytes)
	return ctrler.addDelegatee(staker)
}

func (ctrler *StakeCtrler) addDelegatee(staker *Delegatee) *Delegatee {
	addrKey := types.ToAcctKey(staker.Addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; !ok {
		ctrler.allDelegatees = append(ctrler.allDelegatees, staker)
		ctrler.allDelegateesMap[addrKey] = staker
		return staker
	}
	return nil
}

func (ctrler *StakeCtrler) removeDelegatee(addr types.Address) *Delegatee {
	addrKey := types.ToAcctKey(addr)
	if _, ok := ctrler.allDelegateesMap[addrKey]; ok {
		for i, staker := range ctrler.allDelegatees {
			if bytes.Compare(staker.Addr, addr) == 0 {
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
	if staker, ok := ctrler.allDelegateesMap[addrKey]; ok {
		return staker
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
		staker := ctrler.findDelegatee(ctx.Tx.To)
		if staker == nil && bytes.Compare(ctx.Tx.From, ctx.Tx.To) == 0 {
			// new staker (staking to my self)
			staker = ctrler.addDelegatee(NewDelegatee(ctx.Tx.From, ctx.SenderPubKey))
		} else if staker == nil {
			// there is no staker whose address is ctx.Tx.To
			return xerrors.ErrNotFoundStaker
		}

		s0 := NewStakeWithAmount(ctx.Tx.From, ctx.Tx.To, ctx.Tx.Amount, ctx.Height, ctx.TxHash, ctx.GovRules)
		expectedPower := ctrler.GetTotalPower() + s0.Power
		if expectedPower < 0 || expectedPower > ctx.GovRules.MaxTotalPower() {
			return xerrors.ErrTooManyPower
		}

		if xerr := staker.AppendStake(s0); xerr != nil {
			// Not reachable. AppendStake() does not return error
			ctrler.logger.Error("Not reachable", "error", xerr)
			panic(xerr)
		}

		ctrler.updatedStakes = append(ctrler.updatedStakes, s0)
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

		s0.RefundHeight = ctx.Height + ctx.GovRules.GetRewardDelayBlocks()
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

				s0.RefundHeight = ctx.Height + ctx.GovRules.GetRewardDelayBlocks()

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

func (ctrler *StakeCtrler) Apply(ctx *trxs.TrxContext) error {
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
	for _, staker := range ctrler.allDelegatees {
		if bytes.Compare(staker.Addr, feeOwner) == 0 && fee.Sign() > 0 {
			staker.ApplyFeeReward(fee)
		}

		// reward does not include fee
		reward = reward.Add(reward, staker.ApplyReward())
	}
	return reward
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

func (ctrler *StakeCtrler) GetLastValidatorCnt() int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return len(ctrler.lastValidators)
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

func (ctrler *StakeCtrler) UpdateValidators(maxVals int) []tmtypes.ValidatorUpdate {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	sort.Sort(powerOrderedStakeSets(ctrler.allDelegatees)) // sort by power

	var newValidators ValidatorInfoList
	n := libs.MIN(len(ctrler.allDelegatees), maxVals)
	for i := 0; i < n; i++ {
		staker := ctrler.allDelegatees[i]
		newValidators = append(newValidators, &ValidatorInfo{
			Address: staker.Addr,
			PubKey:  staker.PubKey,
			Power:   staker.TotalPower,
		})
	}
	sort.Sort(newValidators) // sort by address

	ret := validatorUpdates(ctrler.lastValidators, newValidators)
	ctrler.lastValidators = newValidators

	return ret
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

	batch := ctrler.frozenStakesDB.NewBatch()
	defer batch.Close()

	for _, s := range ctrler.delFrozenStakes {
		if err := batch.Delete(s.TxHash); err != nil {
			return nil, -1, err
		}
	}
	ctrler.allFrozenStakes = ctrler.allFrozenStakes[len(ctrler.delFrozenStakes):]
	ctrler.delFrozenStakes = nil
	for _, s := range ctrler.newFrozenStakes {
		if bz, err := json.Marshal(s); err != nil {
			return nil, -1, xerrors.NewFrom(err)
		} else if err := batch.Set(s.TxHash, bz); err != nil {
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

	if err := batch.Write(); err != nil {
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
var _ types.ILedgerCtrler = (*StakeCtrler)(nil)

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

type powerOrderedStakeSets []*Delegatee

func (vs powerOrderedStakeSets) Len() int {
	return len(vs)
}

// descending order by TotalPower
func (vs powerOrderedStakeSets) Less(i, j int) bool {
	if vs[i].TotalPower != vs[j].TotalPower {
		return vs[i].TotalPower > vs[j].TotalPower
	} else if len(vs[i].Stakes) != len(vs[j].Stakes) {
		return len(vs[i].Stakes) > len(vs[j].Stakes)
	} else if bytes.Compare(vs[i].Addr, vs[j].Addr) > 0 {
		return true
	}
	return false
}

func (vs powerOrderedStakeSets) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (powerOrderedStakeSets)(nil)
