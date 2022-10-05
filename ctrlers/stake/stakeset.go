package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kysee/arcanus/types"
	"math/big"
	"sync"
)

type StakeSet struct {
	Owner       types.Address  `json:"owner"`
	PubKey      types.HexBytes `json:"pub_key"`
	Stakes      []*Stake       `json:"stakes"` // ordered by block height
	TotalPower  int64          `json:"total_vp"`
	TotalAmount *big.Int       `json:"total_amount"`
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

func (sset *StakeSet) CalculatePower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.calculatePower()
}

func (sset *StakeSet) calculatePower() int64 {
	vp := int64(0)
	for _, s := range sset.Stakes {
		vp += s.calculatePower()
	}
	return vp
}

func (sset *StakeSet) CalculateReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.calculateReward()
}

func (sset *StakeSet) calculateReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		reward.Add(reward, s.calculateReward())
	}
	return reward
}

func (sset *StakeSet) CurrentReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	// this should be equal to sset.TotalReward
	//return sset.currentReward()
	return new(big.Int).Set(sset.TotalReward)
}

func (sset *StakeSet) currentReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		reward.Add(reward, s.Reward)
	}
	return reward
}

func (sset *StakeSet) ApplyReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.applyReward()
}

func (sset *StakeSet) ApplyFeeReward(fee *big.Int) {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	sset.FeeReward = new(big.Int).Add(sset.FeeReward, fee)
}

func (sset *StakeSet) applyReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		reward = new(big.Int).Add(reward, s.applyReward())
	}
	sset.TotalReward = new(big.Int).Add(sset.TotalReward, reward)
	return reward
}
