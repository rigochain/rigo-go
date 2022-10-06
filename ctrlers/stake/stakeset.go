package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kysee/arcanus/types"
	"math/big"
	"sort"
	"sync"
)

type stakeList []*Stake

func (slst stakeList) Len() int {
	return len(slst)
}

func (slst stakeList) Less(i, j int) bool {
	return slst[i].StartHeight < slst[j].StartHeight
}

func (slst stakeList) Swap(i, j int) {
	slst[i], slst[j] = slst[j], slst[i]
}

var _ sort.Interface = (stakeList)(nil)

type StakeSet struct {
	Owner       types.Address  `json:"owner"`
	PubKey      types.HexBytes `json:"pub_key"`
	Stakes      stakeList      `json:"stakes"` // ordered by block height
	TotalAmount *big.Int       `json:"total_amount"`
	TotalPower  int64          `json:"total_power"`
	TotalReward *big.Int       `json:"total_reward"`
	FeeReward   *big.Int       `json:"fee_reward"`

	mtx sync.RWMutex
}

func NewStakeSet(addr types.Address, pubKey types.HexBytes) *StakeSet {
	return &StakeSet{
		Owner:       addr,
		PubKey:      pubKey,
		TotalPower:  0,
		TotalAmount: big.NewInt(0),
		TotalReward: big.NewInt(0),
		FeeReward:   big.NewInt(0),
	}
}

func (sset *StakeSet) String() string {
	bz, err := json.MarshalIndent(sset, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}

func (sset *StakeSet) AppendStake(stakes ...*Stake) error {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	_ = sset.appendStake(stakes...)

	return nil
}

func (sset *StakeSet) appendStake(stakes ...*Stake) error {
	sset.Stakes = append(sset.Stakes, stakes...)
	sort.Sort(sset.Stakes)

	for _, s := range stakes {
		sset.TotalPower += s.Power
		sset.TotalAmount = new(big.Int).Add(sset.TotalAmount, s.Amount)
	}
	return nil
}

func (sset *StakeSet) PopStake() *Stake {
	return sset.DelStake(0)
}

func (sset *StakeSet) DelStake(idx int) *Stake {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	if s := sset.delStake(idx); s != nil {
		sset.TotalPower -= s.Power
		sset.TotalAmount = new(big.Int).Sub(sset.TotalAmount, s.Amount)
		sset.TotalReward = new(big.Int).Sub(sset.TotalReward, s.Reward)
		return s
	}
	return nil
}

func (sset *StakeSet) delStake(idx int) *Stake {
	if idx >= len(sset.Stakes) {
		return nil
	} else {
		s := sset.Stakes[idx]
		sset.Stakes = append(sset.Stakes[:idx], sset.Stakes[idx+1:]...)
		return s
	}
}

func (sset *StakeSet) GetStake(idx int) *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.getStake(idx)
}

func (sset *StakeSet) getStake(idx int) *Stake {
	if idx >= len(sset.Stakes) {
		return nil
	}
	return sset.Stakes[idx]
}

func (sset *StakeSet) FirstStake() *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.getStake(0)
}

func (sset *StakeSet) LastStake() *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	idx := len(sset.Stakes) - 1
	return sset.getStake(idx)
}

func (sset *StakeSet) findStakeByAddr(addr types.Address) *Stake {
	for _, s := range sset.Stakes {
		if bytes.Compare(addr, s.Owner) == 0 {
			return s
		}
	}
	return nil
}

func (sset *StakeSet) StakesLen() int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return len(sset.Stakes)
}

func (sset *StakeSet) GetTotalAmount() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.TotalAmount
}

func (sset *StakeSet) SumAmount() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumAmount()
}

func (sset *StakeSet) sumAmount() *big.Int {
	amt := big.NewInt(0)
	for _, s := range sset.Stakes {
		_ = amt.Add(amt, s.Amount)
	}
	return amt
}

func (sset *StakeSet) GetTotalPower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.TotalPower
}

func (sset *StakeSet) SumPower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumPower()
}

func (sset *StakeSet) sumPower() int64 {
	power := int64(0)
	for _, s := range sset.Stakes {
		power += s.Power
	}
	return power
}

func (sset *StakeSet) GetTotalReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return new(big.Int).Set(sset.TotalReward)
}

func (sset *StakeSet) SumReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumReward()
}

func (sset *StakeSet) sumReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		_ = reward.Add(reward, s.Reward)
	}
	return reward
}

func (sset *StakeSet) ApplyReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.applyReward()
}

func (sset *StakeSet) applyReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		reward = new(big.Int).Add(reward, s.applyReward())
	}
	sset.TotalReward = new(big.Int).Add(sset.TotalReward, reward)
	return reward
}

func (sset *StakeSet) ApplyFeeReward(fee *big.Int) {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	sset.FeeReward = new(big.Int).Add(sset.FeeReward, fee)
}
