package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"math/big"
	"sort"
	"sync"
)

type Delegatee struct {
	Addr   account.Address `json:"address"`
	PubKey types.HexBytes  `json:"pubKey"`
	Stakes []*Stake        `json:"stakes"`

	SelfAmount  *big.Int `json:"selfAmount"`
	SelfPower   int64    `json:"selfPower"`
	TotalAmount *big.Int `json:"totalAmount"`
	TotalPower  int64    `json:"totalPower"`

	ReceivedReward *Reward `json:"receivedReward"`

	mtx sync.RWMutex
}

func NewDelegatee(addr account.Address, pubKey types.HexBytes) *Delegatee {
	return &Delegatee{
		Addr:           addr,
		PubKey:         pubKey,
		SelfAmount:     big.NewInt(0),
		SelfPower:      0,
		TotalPower:     0,
		TotalAmount:    big.NewInt(0),
		ReceivedReward: NewReward(pubKey),
	}
}

func (delegatee *Delegatee) String() string {
	bz, err := json.MarshalIndent(delegatee, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}

func (delegatee *Delegatee) AppendStake(stakes ...*Stake) error {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	return delegatee.appendStake(stakes...)
}

func (delegatee *Delegatee) appendStake(stakes ...*Stake) error {
	//stakes[n].To is equal to delegatee.Addr

	delegatee.Stakes = append(delegatee.Stakes, stakes...)

	for _, s := range stakes {
		if s.IsSelfStake() {
			delegatee.SelfPower += s.Power
			delegatee.SelfAmount = new(big.Int).Add(delegatee.SelfAmount, s.Amount)
		}
		delegatee.TotalPower += s.Power
		delegatee.TotalAmount = new(big.Int).Add(delegatee.TotalAmount, s.Amount)
		delegatee.ReceivedReward.AddBlockReward(s.ReceivedReward)
	}
	return nil
}

func (delegatee *Delegatee) DelStake(txhash types.HexBytes) *Stake {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	i, s0 := delegatee.findStake(txhash)
	if i < 0 || s0 == nil {
		return nil
	}

	if s := delegatee.delStakeByIdx(i); s != nil {
		if s.IsSelfStake() {
			delegatee.SelfPower -= s.Power
			delegatee.SelfAmount = new(big.Int).Sub(delegatee.SelfAmount, s.Amount)
		}
		delegatee.TotalPower -= s.Power
		delegatee.TotalAmount = new(big.Int).Sub(delegatee.TotalAmount, s.Amount)

		//delegatee.TotalReward = new(big.Int).Sub(delegatee.TotalReward, s.ReceivedReward)
		delegatee.ReceivedReward.SubBlockReward(s.ReceivedReward)
		return s
	}
	return nil
}

func (delegatee *Delegatee) DelStakeByIdx(idx int) *Stake {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	if s := delegatee.delStakeByIdx(idx); s != nil {
		if s.IsSelfStake() {
			delegatee.SelfPower -= s.Power
			delegatee.SelfAmount = new(big.Int).Sub(delegatee.SelfAmount, s.Amount)
		}
		delegatee.TotalPower -= s.Power
		delegatee.TotalAmount = new(big.Int).Sub(delegatee.TotalAmount, s.Amount)

		//delegatee.TotalReward = new(big.Int).Sub(delegatee.TotalReward, s.ReceivedReward)
		delegatee.ReceivedReward.SubBlockReward(s.ReceivedReward)
		return s
	}
	return nil
}

func (delegatee *Delegatee) delStakeByIdx(idx int) *Stake {
	if idx >= len(delegatee.Stakes) {
		return nil
	} else {
		s := delegatee.Stakes[idx]
		delegatee.Stakes = append(delegatee.Stakes[:idx], delegatee.Stakes[idx+1:]...)
		return s
	}
}

func (delegatee *Delegatee) DelAllStakes() []*Stake {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	stakes := delegatee.Stakes
	delegatee.Stakes = nil

	for _, s := range stakes {
		delegatee.TotalPower -= s.Power
		delegatee.TotalAmount = new(big.Int).Sub(delegatee.TotalAmount, s.Amount)

		//delegatee.TotalReward = new(big.Int).Sub(delegatee.TotalReward, s.ReceivedReward)
		delegatee.ReceivedReward.SubBlockReward(s.ReceivedReward)
	}

	return stakes
}

func (delegatee *Delegatee) GetAllStakes() []*Stake {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	ret := make([]*Stake, len(delegatee.Stakes))
	copy(ret, delegatee.Stakes)
	return ret
}

func (delegatee *Delegatee) GetStake(idx int) *Stake {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.getStake(idx)
}

func (delegatee *Delegatee) getStake(idx int) *Stake {
	if idx >= len(delegatee.Stakes) {
		return nil
	}
	return delegatee.Stakes[idx]
}

func (delegatee *Delegatee) NextStake() *Stake {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.getStake(0)
}

func (delegatee *Delegatee) LastStake() *Stake {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	idx := len(delegatee.Stakes) - 1
	return delegatee.getStake(idx)
}

func (delegatee *Delegatee) FindStake(txhash types.HexBytes) (int, *Stake) {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.findStake(txhash)
}

func (delegatee *Delegatee) findStake(txhash types.HexBytes) (int, *Stake) {
	for i, s := range delegatee.Stakes {
		if bytes.Compare(txhash, s.TxHash) == 0 {
			return i, s
		}
	}
	return -1, nil
}

func (delegatee *Delegatee) StakesOf(addr account.Address) []*Stake {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	var ret []*Stake
	for _, s := range delegatee.Stakes {
		if bytes.Compare(addr, s.From) == 0 {
			ret = append(ret, s)
		}
	}
	return ret
}

func (delegatee *Delegatee) StakesLen() int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return len(delegatee.Stakes)
}

func (delegatee *Delegatee) GetSelfAmount() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.SelfAmount
}

func (delegatee *Delegatee) GetTotalAmount() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.TotalAmount
}

func (delegatee *Delegatee) SumAmountOf(addr account.Address) *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	ret := big.NewInt(0)
	for _, s0 := range delegatee.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret = ret.Add(ret, s0.Amount)
		}
	}
	return ret
}

func (delegatee *Delegatee) SumAmount() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumAmount()
}

func (delegatee *Delegatee) sumAmount() *big.Int {
	amt := big.NewInt(0)
	for _, s := range delegatee.Stakes {
		_ = amt.Add(amt, s.Amount)
	}
	return amt
}

func (delegatee *Delegatee) GetSelfPower() int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.SelfPower
}

func (delegatee *Delegatee) GetTotalPower() int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.TotalPower
}

func (delegatee *Delegatee) SumPowerOf(addr account.Address) int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	ret := int64(0)
	for _, s0 := range delegatee.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret += s0.Power
		}
	}
	return ret
}

func (delegatee *Delegatee) SumPower() int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumPower()
}

func (delegatee *Delegatee) sumPower() int64 {
	power := int64(0)
	for _, s := range delegatee.Stakes {
		power += s.Power
	}
	return power
}

func (delegatee *Delegatee) GetTotalReward() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.ReceivedReward.TotalReward()
}

func (delegatee *Delegatee) SumBlockReward() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumBlockReward()
}

func (delegatee *Delegatee) SumBlockRewardOf(addr account.Address) *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	ret := big.NewInt(0)
	for _, s0 := range delegatee.Stakes {
		if bytes.Compare(s0.From, addr) == 0 {
			ret = ret.Add(ret, s0.ReceivedReward)
		}
	}
	return ret
}

func (delegatee *Delegatee) sumBlockReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range delegatee.Stakes {
		_ = reward.Add(reward, s.ReceivedReward)
	}
	return reward
}

func (delegatee *Delegatee) ApplyBlockReward() *big.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.applyBlockReward()
}

func (delegatee *Delegatee) applyBlockReward() *big.Int {
	reward := big.NewInt(0)
	for _, s := range delegatee.Stakes {
		reward = new(big.Int).Add(reward, s.applyReward())
	}

	delegatee.ReceivedReward.AddBlockReward(reward)
	return reward
}

func (delegatee *Delegatee) ApplyFeeReward(fee *big.Int) {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	delegatee.ReceivedReward.AddFeeReward(fee)
}

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
