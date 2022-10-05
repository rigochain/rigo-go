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
	Owner       types.Address `json:"owner"`
	Amount      *big.Int      `json:"amount"`
	Power       int64         `json:"power"`
	BlockReward *big.Int      `json:"block_reward"`
	Reward      *big.Int      `json:"received_reward"`

	StartHeight int64          `json:"start_height"`
	LastHeight  int64          `json:"last_height"`
	TxHash      types.HexBytes `json:"txhash"`

	mtx sync.RWMutex
}

func NewStakeWithAmount(addr types.Address, amt *big.Int, height int64, txhash types.HexBytes, rules types.IGovRules) *Stake {
	power := rules.AmountToPower(amt)
	blockReward := rules.PowerToReward(power)
	return &Stake{
		Owner:       addr,
		Amount:      amt,
		Power:       power,
		BlockReward: blockReward,
		Reward:      big.NewInt(0),
		StartHeight: height,
		LastHeight:  0,
		TxHash:      txhash,
	}
}

func NewStakeWithPower(addr types.Address, power int64, height int64, txhash types.HexBytes, rules types.IGovRules) *Stake {
	amt := rules.PowerToAmount(power)
	blockReward := rules.PowerToReward(power)
	return &Stake{
		Owner:       addr,
		Amount:      amt,
		Power:       power,
		BlockReward: blockReward,
		Reward:      big.NewInt(0),
		StartHeight: height,
		LastHeight:  0,
		TxHash:      txhash,
	}
}

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
		Power:       s.Power,
		BlockReward: new(big.Int).Set(s.BlockReward),
		Reward:      new(big.Int).Set(s.Reward),
		StartHeight: s.StartHeight,
		LastHeight:  s.LastHeight,
	}
}

//
//func (s *Stake) calculatePower() int64 {
//	s.mtx.RLock()
//	defer s.mtx.RUnlock()
//
//	return types.AmountToPower(s.Amount)
//}
//
//func (s *Stake) calculateReward() *big.Int {
//	return types.PowerToReward(s.calculatePower())
//}

func (s *Stake) applyReward() *big.Int {
	s.Reward = new(big.Int).Add(s.Reward, s.BlockReward)
	return s.BlockReward
}

func (s *Stake) currentReward() *big.Int {
	return s.Reward
}

func (s *Stake) currentBlockReward() *big.Int {
	return s.BlockReward
}

func (s *Stake) String() string {
	bz, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}
