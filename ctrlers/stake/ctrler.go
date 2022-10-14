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
	stakersDB   db.DB
	stakersTree *iavl.MutableTree

	allStakers     StakeSetArray
	allStakersMap  map[types.AcctKey]*StakeSet
	updatedStakers StakeSetArray // removed + updated + newer

	lastValidators ValidatorInfoList

	frozenDB        db.DB
	allFrozenStakes []*Stake
	newFrozenStakes []*Stake
	delFrozenStakes []*Stake

	logger log.Logger
	mtx    sync.RWMutex
}

func NewStakeCtrler(dbDir string, logger log.Logger) (*StakeCtrler, error) {

	// for all stakers
	stakeDB, err := db.NewDB("stake", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}
	stakersTree, err := iavl.NewMutableTree(stakeDB, 1024)
	if err != nil {
		return nil, err
	}
	if _, err := stakersTree.Load(); err != nil {
		return nil, err
	}

	var allStakers StakeSetArray
	allStakersMap := make(map[types.AcctKey]*StakeSet)

	stopped, err := stakersTree.Iterate(func(key []byte, value []byte) bool {
		staker := &StakeSet{}
		if err := json.Unmarshal(value, staker); err != nil {
			logger.Error("Unable to load staker", "address", types.HexBytes(key), "error", err)
			return true
		}

		sumPower := staker.SumPower()
		if staker.TotalPower != sumPower {
			logger.Error("Wrong power", "TotalPower", staker.TotalPower, "Sum of powers of stakes", sumPower)
			return true
		}

		addrKey := types.ToAcctKey(key)
		if stakeSet, ok := allStakersMap[addrKey]; ok {
			logger.Error("Conflict staker", "address", types.HexBytes(key), "already existed", stakeSet)
			return true
		}

		allStakers = append(allStakers, staker)
		allStakersMap[types.ToAcctKey(key)] = staker
		return false
	})
	if err != nil {
		return nil, xerrors.NewFrom(err)
	} else if stopped {
		return nil, xerrors.New("Stop to load stakers tree")
	}

	sort.Sort(allStakers)

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
		stakersDB:       stakeDB,
		stakersTree:     stakersTree,
		allStakers:      allStakers,
		allStakersMap:   allStakersMap,
		frozenDB:        frozenDB,
		allFrozenStakes: allFrozenStakes,
		logger:          logger,
	}
	return ret, nil
}

func (ctrler *StakeCtrler) AddStaker(staker *StakeSet) *StakeSet {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.addStaker(staker)
}

func (ctrler *StakeCtrler) AddStakerWith(addr types.Address, pubKeyBytes types.HexBytes) *StakeSet {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	staker := NewStakeSet(addr, pubKeyBytes)
	return ctrler.addStaker(staker)
}

func (ctrler *StakeCtrler) addStaker(staker *StakeSet) *StakeSet {
	addrKey := types.ToAcctKey(staker.Owner)
	if _, ok := ctrler.allStakersMap[addrKey]; !ok {
		ctrler.allStakers = append(ctrler.allStakers, staker)
		ctrler.allStakersMap[addrKey] = staker
		return staker
	}
	return nil
}

func (ctrler *StakeCtrler) removeStaker(addr types.Address) *StakeSet {
	addrKey := types.ToAcctKey(addr)
	if _, ok := ctrler.allStakersMap[addrKey]; ok {
		for i, staker := range ctrler.allStakers {
			if bytes.Compare(staker.Owner, addr) == 0 {
				ctrler.allStakers = append(ctrler.allStakers[:i], ctrler.allStakers[i+1:]...)
				delete(ctrler.allStakersMap, addrKey)
				return staker
			}
		}

		ctrler.logger.Error("not same allStakers and allStakersMap)", "address", addr)
	}
	return nil
}

func (ctrler *StakeCtrler) GetStaker(idx int) *StakeSet {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if idx >= len(ctrler.allStakers) {
		return nil
	}

	return ctrler.allStakers[idx]
}

func (ctrler *StakeCtrler) StakersLen() int {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return len(ctrler.allStakers)
}

func (ctrler *StakeCtrler) FindStaker(addr types.Address) *StakeSet {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findStaker(addr)
}

func (ctrler *StakeCtrler) findStaker(addr types.Address) *StakeSet {
	addrKey := types.ToAcctKey(addr)
	if staker, ok := ctrler.allStakersMap[addrKey]; ok {
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
		staker := ctrler.findStaker(ctx.Tx.To)
		if staker == nil && bytes.Compare(ctx.Tx.From, ctx.Tx.To) == 0 {
			// new staker (staking to my self)
			staker = ctrler.addStaker(NewStakeSet(ctx.Tx.From, ctx.SenderPubKey))
		} else if staker == nil {
			// there is no staker whose address is ctx.Tx.To
			return xerrors.ErrNotFoundStaker
		}

		if xerr := staker.AppendStake(NewStakeWithAmount(ctx.Tx.From, ctx.Tx.Amount, ctx.Height, ctx.TxHash, ctx.GovRules)); xerr != nil {
			// Not reachable. AppendStake() does not return error
			ctrler.logger.Error("Not reachable", "error", xerr)
			panic(xerr)
		}

		ctrler.updatedStakers = append(ctrler.updatedStakers, staker)
	}

	return nil
}

func (ctrler *StakeCtrler) applyUnstaking(ctx *trxs.TrxContext) error {
	if ctx.Exec {
		staker := ctrler.findStaker(ctx.Tx.To)
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
			_ = ctrler.removeStaker(staker.Owner)
		}
		ctrler.updatedStakers = append(ctrler.updatedStakers, staker)
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
		return ctrler.applyUnstaking(ctx)
	default:
		return xerrors.New("unknown transaction type")
	}
}

func (ctrler *StakeCtrler) ApplyReward(feeOwner types.Address, fee *big.Int) *big.Int {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	reward := big.NewInt(0)
	for _, staker := range ctrler.allStakers {
		if bytes.Compare(staker.Owner, feeOwner) == 0 && fee.Sign() > 0 {
			staker.ApplyFeeReward(fee)
		}

		// reward does not include fee
		reward = reward.Add(reward, staker.ApplyReward())
	}
	return reward
}

func (ctrler *StakeCtrler) ProcessFrozenStakesAt(height int64, acctFinder types.IAccountFinder) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// ctrler.allFrozenStakes is ordered by RefundHeight
	for _, s0 := range ctrler.allFrozenStakes {
		if height >= s0.RefundHeight {
			acct := acctFinder.FindAccount(s0.Owner, true)
			if acct != nil {
				return xerrors.ErrNotFoundAccount
			}

			if xerr := acct.AddBalance(new(big.Int).Add(s0.Amount, s0.Reward)); xerr != nil {
				return xerr
			}
			ctrler.delFrozenStakes = append(ctrler.delFrozenStakes, s0)
		} else {
			break
		}
	}
	return nil
}

func (ctrler *StakeCtrler) UpdateValidators(govRules types.IGovRules) []tmtypes.ValidatorUpdate {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	sort.Sort(ctrler.allStakers)

	var newValidators ValidatorInfoList
	n := libs.MIN(len(ctrler.allStakers), int(govRules.MaxValidatorCount()))
	for i := 0; i < n; i++ {
		staker := ctrler.allStakers[i]
		newValidators = append(newValidators, &ValidatorInfo{
			Address: staker.Owner,
			PubKey:  staker.PubKey,
			Power:   staker.TotalPower,
		})
	}
	sort.Sort(newValidators)

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
			// lastValidator[i] is updated to newValidators[j]
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
			i++
			j++
		} else { // ret > 0
			// newValidators[j] is added
			valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
			j++
		}
	}

	for ; i < len(existing); i++ {
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(existing[i].PubKey, 0, "secp256k1"))
	}
	for ; j < len(newers); i++ {
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(newers[j].PubKey, newers[j].Power, "secp256k1"))
	}

	return valUpdates
}

func (ctrler *StakeCtrler) Commit() ([]byte, int64, error) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// todo: Commit() should be atomic operation

	batch := ctrler.frozenDB.NewBatch()
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
			return nil, -1, err
		} else if err := batch.Set(s.TxHash, bz); err != nil {
			return nil, -1, err
		}
	}
	if err := batch.Write(); err != nil {
		return nil, -1, err
	}

	ctrler.allFrozenStakes = append(ctrler.allFrozenStakes, ctrler.newFrozenStakes...)
	ctrler.newFrozenStakes = nil
	sort.Sort(refundHeightOrder(ctrler.allFrozenStakes))

	for _, staker := range ctrler.updatedStakers {
		if staker.GetTotalPower() == 0 {
			// removed
			if _, _, err := ctrler.stakersTree.Remove(staker.Owner); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			}
		} else {
			// new or updated
			if bz, err := json.Marshal(staker); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			} else if _, err := ctrler.stakersTree.Set(staker.Owner, bz); err != nil {
				return nil, -1, xerrors.NewFrom(err)
			}
		}
	}
	ctrler.updatedStakers = nil

	return ctrler.stakersTree.SaveVersion()
}

func (ctrler *StakeCtrler) Close() error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.stakersDB != nil {
		if err := ctrler.stakersDB.Close(); err != nil {
			return nil
		}
	}
	if ctrler.frozenDB != nil {
		if err := ctrler.frozenDB.Close(); err != nil {
			return err
		}
	}

	ctrler.stakersDB = nil
	ctrler.stakersTree = nil
	ctrler.frozenDB = nil
	return nil
}

var _ trxs.ITrxHandler = (*StakeCtrler)(nil)
var _ types.ILedgerCtrler = (*StakeCtrler)(nil)

type StakeSetArray []*StakeSet

func (vs StakeSetArray) Len() int {
	return len(vs)
}

// descending order by TotalPower
func (vs StakeSetArray) Less(i, j int) bool {
	return vs[i].TotalPower > vs[j].TotalPower
}

func (vs StakeSetArray) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (StakeSetArray)(nil)

func (vs StakeSetArray) SumTotalAmount() *big.Int {
	var amt *big.Int
	for _, val := range vs {
		amt = new(big.Int).Add(amt, val.TotalAmount)
	}
	return amt
}

func (vs StakeSetArray) SumTotalReward() *big.Int {
	var reward *big.Int
	for _, val := range vs {
		reward = new(big.Int).Add(reward, val.TotalReward)
	}
	return reward
}

func (vs StakeSetArray) SumTotalFeeReward() *big.Int {
	var fee *big.Int
	for _, val := range vs {
		fee = new(big.Int).Add(fee, val.FeeReward)
	}
	return fee
}

func (vs StakeSetArray) SumTotalPower() int64 {
	power := int64(0)
	for _, val := range vs {
		power += val.TotalPower
	}
	return power
}
