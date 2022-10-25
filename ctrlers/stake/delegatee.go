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

type startHeightOrder []*Stake

func (slst startHeightOrder) Len() int {
	return len(slst)
}

// ascending order
func (slst startHeightOrder) Less(i, j int) bool {
	return slst[i].StartHeight < slst[j].StartHeight
}

func (slst startHeightOrder) Swap(i, j int) {
	slst[i], slst[j] = slst[j], slst[i]
}

var _ sort.Interface = (startHeightOrder)(nil)

type refundHeightOrder []*Stake

func (slst refundHeightOrder) Len() int {
	return len(slst)
}

// ascending order
func (slst refundHeightOrder) Less(i, j int) bool {
	return slst[i].RefundHeight < slst[j].RefundHeight
}

func (slst refundHeightOrder) Swap(i, j int) {
	slst[i], slst[j] = slst[j], slst[i]
}

var _ sort.Interface = (refundHeightOrder)(nil)

type Delegatee struct {
	Addr   types.Address  `json:"address"`
	PubKey types.HexBytes `json:"pubKey"`
	Stakes []*Stake       `json:"stakes"`

	SelfAmount  *big.Int `json:"selfAmount"`
	SelfPower   int64    `json:"selfPower"`
	TotalAmount *big.Int `json:"totalAmount"`
	TotalPower  int64    `json:"totalPower"`
	TotalReward *big.Int `json:"totalReward"`
	FeeReward   *big.Int `json:"feeReward"`

	mtx sync.RWMutex
}

func NewDelegatee(addr types.Address, pubKey types.HexBytes) *Delegatee {
	return &Delegatee{
		Addr:        addr,
		PubKey:      pubKey,
		SelfAmount:  big.NewInt(0),
		SelfPower:   0,
		TotalPower:  0,
		TotalAmount: big.NewInt(0),
		TotalReward: big.NewInt(0),
		FeeReward:   big.NewInt(0),
	}
}

func (sset *Delegatee) String() string {
	bz, err := json.MarshalIndent(sset, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}

func (sset *Delegatee) AppendStake(stakes ...*Stake) error {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	return sset.appendStake(stakes...)
}

func (sset *Delegatee) appendStake(stakes ...*Stake) error {
	//stakes[n].To is equal to sset.Addr

	sset.Stakes = append(sset.Stakes, stakes...)

	for _, s := range stakes {
		if s.IsSelfStake() {
			sset.SelfPower += s.Power
			sset.SelfAmount = new(big.Int).Add(sset.SelfAmount, s.Amount)
		}
		sset.TotalPower += s.Power
		sset.TotalAmount = new(big.Int).Add(sset.TotalAmount, s.Amount)
	}
	return nil
}

func (sset *Delegatee) DelStake(txhash types.HexBytes) *Stake {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	i, s0 := sset.findStake(txhash)
	if i < 0 || s0 == nil {
		return nil
	}

	if s := sset.delStakeByIdx(i); s != nil {
		if s.IsSelfStake() {
			sset.SelfPower -= s.Power
			sset.SelfAmount = new(big.Int).Sub(sset.SelfAmount, s.Amount)
		}
		sset.TotalPower -= s.Power
		sset.TotalAmount = new(big.Int).Sub(sset.TotalAmount, s.Amount)
		sset.TotalReward = new(big.Int).Sub(sset.TotalReward, s.Reward)
		return s
	}
	return nil
}

func (sset *Delegatee) DelStakeByIdx(idx int) *Stake {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	if s := sset.delStakeByIdx(idx); s != nil {
		if s.IsSelfStake() {
			sset.SelfPower -= s.Power
			sset.SelfAmount = new(big.Int).Sub(sset.SelfAmount, s.Amount)
		}
		sset.TotalPower -= s.Power
		sset.TotalAmount = new(big.Int).Sub(sset.TotalAmount, s.Amount)
		sset.TotalReward = new(big.Int).Sub(sset.TotalReward, s.Reward)
		return s
	}
	return nil
}

func (sset *Delegatee) delStakeByIdx(idx int) *Stake {
	if idx >= len(sset.Stakes) {
		return nil
	} else {
		s := sset.Stakes[idx]
		sset.Stakes = append(sset.Stakes[:idx], sset.Stakes[idx+1:]...)
		return s
	}
}

func (sset *Delegatee) DelAllStakes() []*Stake {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	stakes := sset.Stakes
	sset.Stakes = nil

	for _, s := range stakes {
		sset.TotalPower -= s.Power
		sset.TotalAmount = new(big.Int).Sub(sset.TotalAmount, s.Amount)
		sset.TotalReward = new(big.Int).Sub(sset.TotalReward, s.Reward)
	}

	return stakes
}

func (sset *Delegatee) GetStake(idx int) *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.getStake(idx)
}

func (sset *Delegatee) getStake(idx int) *Stake {
	if idx >= len(sset.Stakes) {
		return nil
	}
	return sset.Stakes[idx]
}

func (sset *Delegatee) NextStake() *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.getStake(0)
}

func (sset *Delegatee) LastStake() *Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	idx := len(sset.Stakes) - 1
	return sset.getStake(idx)
}

func (sset *Delegatee) FindStake(txhash types.HexBytes) (int, *Stake) {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.findStake(txhash)
}

func (sset *Delegatee) findStake(txhash types.HexBytes) (int, *Stake) {
	for i, s := range sset.Stakes {
		if bytes.Compare(txhash, s.TxHash) == 0 {
			return i, s
		}
	}
	return -1, nil
}

func (sset *Delegatee) StakesOf(addr types.Address) []*Stake {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	var ret []*Stake
	for _, s := range sset.Stakes {
		if bytes.Compare(addr, s.From) == 0 {
			ret = append(ret, s)
		}
	}
	return ret
}

func (sset *Delegatee) StakesLen() int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return len(sset.Stakes)
}

func (sset *Delegatee) GetSelfAmount() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.SelfAmount
}

func (sset *Delegatee) GetTotalAmount() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.TotalAmount
}

func (sset *Delegatee) SumAmountOf(addr types.Address) *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	ret := big.NewInt(0)
	for _, s0 := range sset.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret = ret.Add(ret, s0.Amount)
		}
	}
	return ret
}

func (sset *Delegatee) SumAmount() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumAmount()
}

func (sset *Delegatee) sumAmount() *big.Int {
	amt := big.NewInt(0)
	for _, s := range sset.Stakes {
		_ = amt.Add(amt, s.Amount)
	}
	return amt
}

func (sset *Delegatee) GetSelfPower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.SelfPower
}

func (sset *Delegatee) GetTotalPower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.TotalPower
}

func (sset *Delegatee) SumPowerOf(addr types.Address) int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	ret := int64(0)
	for _, s0 := range sset.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret += s0.Power
		}
	}
	return ret
}

func (sset *Delegatee) SumPower() int64 {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumPower()
}

func (sset *Delegatee) sumPower() int64 {
	power := int64(0)
	for _, s := range sset.Stakes {
		power += s.Power
	}
	return power
}

func (sset *Delegatee) GetTotalReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return new(big.Int).Set(sset.TotalReward)
}

func (sset *Delegatee) SumRewardOf(addr types.Address) *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	ret := big.NewInt(0)
	for _, s0 := range sset.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret = ret.Add(ret, s0.Reward)
		}
	}
	return ret
}

func (sset *Delegatee) SumReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.sumReward()
}

func (sset *Delegatee) sumReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		_ = reward.Add(reward, s.Reward)
	}
	return reward
}

func (sset *Delegatee) ApplyReward() *big.Int {
	sset.mtx.RLock()
	defer sset.mtx.RUnlock()

	return sset.applyReward()
}

func (sset *Delegatee) applyReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range sset.Stakes {
		reward = new(big.Int).Add(reward, s.applyReward())
	}
	sset.TotalReward = new(big.Int).Add(sset.TotalReward, reward)
	return reward
}

func (sset *Delegatee) ApplyFeeReward(fee *big.Int) {
	sset.mtx.Lock()
	defer sset.mtx.Unlock()

	sset.FeeReward = new(big.Int).Add(sset.FeeReward, fee)
}
