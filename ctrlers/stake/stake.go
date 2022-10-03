package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kysee/arcanus/types"
	"math/big"
	"sync"
)

type Stake struct {
	Owner  types.Address `json:"owner"`
	Power  int64         `json:"voting_power"`
	Amount *big.Int      `json:"amount"`
	Reward *big.Int      `json:"reward"`

	StartHeight int64          `json:"start_height"`
	LastHeight  int64          `json:"last_height"`
	TxHash      types.HexBytes `json:"txhash"`

	mtx sync.RWMutex
}

func NewStake(addr types.Address, amt *big.Int, height int64, txhash types.HexBytes) *Stake {
	return &Stake{
		Owner:       addr,
		Power:       types.AmountToPower(amt),
		Amount:      amt,
		Reward:      big.NewInt(0),
		StartHeight: height,
		LastHeight:  0,
		TxHash:      txhash,
	}
}

//func (s *Stake) IncreaseAmount(amt *big.Int) {
//	s.mtx.Lock()
//	defer s.mtx.Unlock()
//
//	s.Amount = new(big.Int).Add(s.Amount, amt)
//	s.Power += 0 // todo: add power
//}
//
//func (s *Stake) DecreaseAmount(amt *big.Int) {
//	s.mtx.Lock()
//	defer s.mtx.Unlock()
//
//	if s.Amount.Cmp(amt) < 0 {
//		return
//	}
//
//	s.Amount = new(big.Int).Sub(s.Amount, amt)
//	s.Power -= 0 // todo: sub power
//}

func (s *Stake) Equal(o *Stake) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return bytes.Compare(s.Owner, o.Owner) == 0 &&
		bytes.Compare(s.TxHash, o.TxHash) == 0 &&
		s.StartHeight == o.StartHeight &&
		s.Amount.Cmp(o.Amount) == 0
}

func (s *Stake) Copy() *Stake {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return &Stake{
		Owner:       s.Owner,
		Amount:      new(big.Int).Set(s.Amount),
		Reward:      new(big.Int).Set(s.Reward),
		StartHeight: s.StartHeight,
		LastHeight:  s.LastHeight,
	}
}

func (s *Stake) calculatePower() int64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return types.AmountToPower(s.Amount)
}

func (s *Stake) calculateReward() *big.Int {
	return types.PowerToReward(s.calculatePower())
}

func (s *Stake) applyReward() *big.Int {
	_rew := s.calculateReward()
	s.Reward = new(big.Int).Add(s.Reward, _rew)
	return _rew
}

func (s *Stake) currentReward() *big.Int {
	return s.Reward
}

func (s *Stake) String() string {
	bz, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}
