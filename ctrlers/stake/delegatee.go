package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"sort"
	"sync"
)

type Delegatee struct {
	Addr   types.Address   `json:"address"`
	PubKey bytes2.HexBytes `json:"pubKey"`

	SelfAmount  *uint256.Int `json:"selfAmount"`
	SelfPower   int64        `json:"selfPower,string"`
	TotalAmount *uint256.Int `json:"totalAmount"`
	TotalPower  int64        `json:"totalPower,string"`

	RewardAmount *uint256.Int `json:"rewardAmount"`
	Stakes       []*Stake     `json:"stakes"`

	mtx sync.RWMutex
}

func (delegatee *Delegatee) Key() ledger.LedgerKey {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return ledger.ToLedgerKey(delegatee.Addr)
}

func (delegatee *Delegatee) Encode() ([]byte, xerrors.XError) {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	if bz, err := json.Marshal(delegatee); err != nil {
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (delegatee *Delegatee) Decode(d []byte) xerrors.XError {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	if err := json.Unmarshal(d, delegatee); err != nil {
		return xerrors.From(err)
	}
	return nil
}

var _ ledger.ILedgerItem = (*Delegatee)(nil)

func NewDelegatee(addr types.Address, pubKey bytes2.HexBytes) *Delegatee {
	return &Delegatee{
		Addr:         addr,
		PubKey:       pubKey,
		SelfAmount:   uint256.NewInt(0),
		SelfPower:    0,
		TotalPower:   0,
		TotalAmount:  uint256.NewInt(0),
		RewardAmount: uint256.NewInt(0),
	}
}

func (delegatee *Delegatee) AddStake(stakes ...*Stake) xerrors.XError {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	return delegatee.addStake(stakes...)
}

func (delegatee *Delegatee) addStake(stakes ...*Stake) xerrors.XError {

	delegatee.Stakes = append(delegatee.Stakes, stakes...)

	for _, s := range stakes {
		if s.IsSelfStake() {
			delegatee.SelfPower += s.Power
			_ = delegatee.SelfAmount.Add(delegatee.SelfAmount, s.Amount)
		}
		delegatee.TotalPower += s.Power
		_ = delegatee.TotalAmount.Add(delegatee.TotalAmount, s.Amount)
		_ = delegatee.RewardAmount.Add(delegatee.RewardAmount, s.ReceivedReward)
	}
	return nil
}

func (delegatee *Delegatee) DelStake(txhash bytes2.HexBytes) *Stake {
	delegatee.mtx.Lock()
	defer delegatee.mtx.Unlock()

	i, s0 := delegatee.findStake(txhash)
	if i < 0 || s0 == nil {
		return nil
	}

	if s := delegatee.delStakeByIdx(i); s != nil {
		if s.IsSelfStake() {
			delegatee.SelfPower -= s.Power
			_ = delegatee.SelfAmount.Sub(delegatee.SelfAmount, s.Amount)
		}
		delegatee.TotalPower -= s.Power
		_ = delegatee.TotalAmount.Sub(delegatee.TotalAmount, s.Amount)
		_ = delegatee.RewardAmount.Sub(delegatee.RewardAmount, s.ReceivedReward)
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
		_ = delegatee.TotalAmount.Sub(delegatee.TotalAmount, s.Amount)
		_ = delegatee.RewardAmount.Sub(delegatee.RewardAmount, s.ReceivedReward)
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

//	func (delegatee *Delegatee) NextStake() *Stake {
//		delegatee.mtx.RLock()
//		defer delegatee.mtx.RUnlock()
//
//		return delegatee.getStake(0)
//	}
//
//	func (delegatee *Delegatee) LastStake() *Stake {
//		delegatee.mtx.RLock()
//		defer delegatee.mtx.RUnlock()
//
//		idx := len(delegatee.Stakes) - 1
//		return delegatee.getStake(idx)
//	}

func (delegatee *Delegatee) FindStake(txhash bytes2.HexBytes) (int, *Stake) {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.findStake(txhash)
}

func (delegatee *Delegatee) findStake(txhash bytes2.HexBytes) (int, *Stake) {
	for i, s := range delegatee.Stakes {
		if bytes.Compare(txhash, s.TxHash) == 0 {
			return i, s
		}
	}
	return -1, nil
}

func (delegatee *Delegatee) StakesOf(addr types.Address) []*Stake {
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

func (delegatee *Delegatee) GetSelfAmount() *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.SelfAmount.Clone()
}

func (delegatee *Delegatee) GetTotalAmount() *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.TotalAmount.Clone()
}

func (delegatee *Delegatee) SumAmount() *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumAmountOf(nil)
}

func (delegatee *Delegatee) SumAmountOf(addr types.Address) *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumAmountOf(addr)
}

func (delegatee *Delegatee) sumAmountOf(addr types.Address) *uint256.Int {
	amt := uint256.NewInt(0)
	for _, s0 := range delegatee.Stakes {
		if addr == nil || bytes.Compare(s0.From, addr) == 0 {
			_ = amt.Add(amt, s0.Amount)
		}
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

func (delegatee *Delegatee) SumPower() int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumPowerOf(nil)
}

func (delegatee *Delegatee) SumPowerOf(addr types.Address) int64 {
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

func (delegatee *Delegatee) sumPowerOf(addr types.Address) int64 {
	power := int64(0)
	for _, s := range delegatee.Stakes {
		if addr == nil || bytes.Compare(addr, s.From) == 0 {
			power += s.Power
		}
	}
	return power
}

func (delegatee *Delegatee) SelfStakeRatio(added int64) int64 {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return (delegatee.SelfPower * int64(100)) / (delegatee.TotalPower + added)
}

func (delegatee *Delegatee) GetRewardAmount() *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.RewardAmount.Clone()
}

func (delegatee *Delegatee) SumBlockReward() *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumBlockRewardOf(nil)
}

func (delegatee *Delegatee) SumBlockRewardOf(addr types.Address) *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.sumBlockRewardOf(addr)
}

func (delegatee *Delegatee) sumBlockRewardOf(addr types.Address) *uint256.Int {
	reward := uint256.NewInt(0)
	for _, s := range delegatee.Stakes {
		if addr == nil || bytes.Compare(addr, s.From) == 0 {
			_ = reward.Add(reward, s.ReceivedReward)
		}
	}
	return reward
}

func (delegatee *Delegatee) DoReward(height int64) *uint256.Int {
	delegatee.mtx.RLock()
	defer delegatee.mtx.RUnlock()

	return delegatee.doBlockReward(height)
}

func (delegatee *Delegatee) doBlockReward(height int64) *uint256.Int {
	reward := uint256.NewInt(0)
	for _, s := range delegatee.Stakes {

		// issue #29
		// `doBlockReward` is called after running `execStaking/execUnstaking`.
		// So the `delegatee` has new stakes at now.
		// Rewarding should be given only to old stakes.
		if s.StartHeight <= height {
			_ = reward.Add(reward, s.applyReward(height))
		}
	}
	_ = delegatee.RewardAmount.Add(delegatee.RewardAmount, reward)
	return reward
}

func (delegatee *Delegatee) String() string {
	bz, err := json.MarshalIndent(delegatee, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}

//
// DelegateeArray

type DelegateeArray []*Delegatee

func (vs DelegateeArray) SumTotalAmount() *uint256.Int {
	amt := uint256.NewInt(0)
	for _, val := range vs {
		_ = amt.Add(amt, val.TotalAmount)
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

func (vs DelegateeArray) SumTotalReward() *uint256.Int {
	reward := uint256.NewInt(0)
	for _, val := range vs {
		_ = reward.Add(reward, val.GetRewardAmount())
	}
	return reward
}

func (vs DelegateeArray) SumBlockReward() *uint256.Int {
	reward := uint256.NewInt(0)
	for _, val := range vs {
		_ = reward.Add(reward, val.RewardAmount)
	}
	return reward
}

type PowerOrderDelegatees []*Delegatee

func (vs PowerOrderDelegatees) Len() int {
	return len(vs)
}

// descending order by TotalPower
func (vs PowerOrderDelegatees) Less(i, j int) bool {
	if vs[i].TotalPower != vs[j].TotalPower {
		return vs[i].TotalPower > vs[j].TotalPower
	}
	if len(vs[i].Stakes) != len(vs[j].Stakes) {
		return len(vs[i].Stakes) > len(vs[j].Stakes)
	}
	if bytes.Compare(vs[i].Addr, vs[j].Addr) > 0 {
		return true
	}
	return false
}

func (vs PowerOrderDelegatees) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (PowerOrderDelegatees)(nil)

type AddressOrderDelegatees []*Delegatee

func (vs AddressOrderDelegatees) Len() int {
	return len(vs)
}

// ascending order by address
func (vs AddressOrderDelegatees) Less(i, j int) bool {
	return bytes.Compare(vs[i].Addr, vs[j].Addr) < 0
}

func (vs AddressOrderDelegatees) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

var _ sort.Interface = (AddressOrderDelegatees)(nil)
