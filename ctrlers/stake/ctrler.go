package stake

import (
	"bytes"
	"github.com/cosmos/iavl"
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

const (
	GOV_MAX_VALIDATORS = 21
)

type StakeCtrler struct {
	stakeDB       db.DB
	stakersTree   *iavl.MutableTree
	allStakers    StakeSetArray
	allStakersMap map[types.AcctKey]*StakeSet

	updatedStakers StakeSetArray
	removedStakers StakeSetArray

	logger log.Logger
	mtx    sync.RWMutex
}

func NewStakeCtrler(dbDir string, logger log.Logger) (*StakeCtrler, error) {
	// for all stakers
	stakeDB, err := db.NewDB("stake", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}
	stakersTree, err := iavl.NewMutableTree(stakeDB, 128)
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
		return nil, xerrors.Wrap(err)
	} else if stopped {
		return nil, xerrors.New("Stop to load stakers tree")
	}

	sort.Sort(allStakers)

	ret := &StakeCtrler{
		stakeDB:       stakeDB,
		stakersTree:   stakersTree,
		allStakers:    allStakers,
		allStakersMap: allStakersMap,
		//valDB:       valDB,
		//validators:  validators,
		logger: logger,
	}
	return ret, nil
}

func (ctrler *StakeCtrler) AddStake(staker *StakeSet) *StakeSet {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.addStaker(staker)
}

func (ctrler *StakeCtrler) AddStakerWith(addr types.Address, pubKey types.HexBytes) *StakeSet {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	staker := NewStakeSet(addr, pubKey)
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

		ctrler.logger.Error("Not same array and map of stakers", "address", addr)
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
		addrKey := types.ToAcctKey(ctx.Tx.From)

		staker, ok := ctrler.allStakersMap[addrKey]
		if !ok {
			// new staker
			staker = ctrler.addStaker(NewStakeSet(ctx.Tx.From, ctx.SenderPubKey))
		}

		if xerr := staker.AppendStake(NewStakeWithAmount(ctx.Tx.From, ctx.Tx.Amount, ctx.Height, ctx.TxHash)); xerr != nil {
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
		staker := ctrler.findStaker(ctx.Tx.From)
		if staker == nil {
			xerrors.New("not found stake. can not do unstaking")
		}

		damt := ctx.Tx.Amount
		if damt.Cmp(staker.TotalAmount) > 0 {
			return xerrors.New("invalid unstaking amount")
		}

		for s := staker.FirstStake(); s != nil; s = staker.FirstStake() {
			if damt.Cmp(s.Amount) >= 0 {
				// damt >= s.Amount
				// remove it
				s = staker.PopStake()
				damt = new(big.Int).Sub(damt, s.Amount)

				// 's' should be handled as frozen stake.

			} else {
				//
				// Now, partial un-staking is not supported.
				// todo: implement partial unstaking

				//s.DecreaseAmount(damt)

				break
			}
		}

		// staker.TotalPower, staker.TotalAmount, staker.Stakes.Len() is related...
		// todo: these variables (TotalPower, TotalAmount, Stakes length) should be checked.
		if staker.TotalPower == 0 {
			_staker := ctrler.removeStaker(staker.Owner)
			ctrler.removedStakers = append(ctrler.removedStakers, _staker)
		} else {
			ctrler.updatedStakers = append(ctrler.updatedStakers, staker)
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
			//reward = reward.Add(reward, fee)
		}
		reward = reward.Add(reward, staker.ApplyReward())
	}
	return reward
}

func (ctrler *StakeCtrler) UpdateValidators() []tmtypes.ValidatorUpdate {
	sort.Sort(ctrler.allStakers)

	var valUpdates []tmtypes.ValidatorUpdate

	for _, staker := range ctrler.updatedStakers {
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(staker.PubKey, staker.TotalPower, "secp256k1"))
	}
	for _, staker := range ctrler.removedStakers {
		valUpdates = append(valUpdates, tmtypes.UpdateValidator(staker.PubKey, staker.TotalPower, "secp256k1"))
	}
	return valUpdates
}

func (ctrler *StakeCtrler) Commit() ([]byte, int64, error) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	for _, s := range ctrler.updatedStakers {
		if bz, err := json.Marshal(s); err != nil {
			return nil, -1, xerrors.Wrap(err)
		} else if _, err := ctrler.stakersTree.Set(s.Owner, bz); err != nil {
			return nil, -1, xerrors.Wrap(err)
		}
	}

	for _, s := range ctrler.removedStakers {
		if _, _, err := ctrler.stakersTree.Remove(s.Owner); err != nil {
			return nil, -1, xerrors.Wrap(err)
		}
	}

	ctrler.updatedStakers = nil
	ctrler.removedStakers = nil
	return ctrler.stakersTree.SaveVersion()
}

func (ctrler *StakeCtrler) Close() error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.stakeDB != nil {
		if err := ctrler.stakeDB.Close(); err != nil {
			return nil
		}
	}

	ctrler.stakeDB = nil
	ctrler.stakersTree = nil
	return nil
}

var _ trxs.ITrxHandler = (*StakeCtrler)(nil)
var _ types.ILedgerCtrler = (*StakeCtrler)(nil)

type StakeSetArray []*StakeSet

func (vs StakeSetArray) Len() int {
	return len(vs)
}

func (vs StakeSetArray) Less(i, j int) bool {
	return vs[i].TotalPower > vs[j].TotalPower
}

func (vs StakeSetArray) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (StakeSetArray)(nil)

func (vs StakeSetArray) TotalAmount() *big.Int {
	var amt *big.Int
	for _, val := range vs {
		amt = new(big.Int).Add(amt, val.TotalAmount)
	}
	return amt
}

func (vs StakeSetArray) TotalReward() *big.Int {
	var reward *big.Int
	for _, val := range vs {
		reward = new(big.Int).Add(reward, val.TotalReward)
	}
	return reward
}

func (vs StakeSetArray) TotalFeeReward() *big.Int {
	var fee *big.Int
	for _, val := range vs {
		fee = new(big.Int).Add(fee, val.FeeReward)
	}
	return fee
}

func (vs StakeSetArray) TotalPower() int64 {
	power := int64(0)
	for _, val := range vs {
		power += val.TotalPower
	}
	return power
}
