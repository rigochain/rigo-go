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
	From        types.Address `json:"owner"`
	To          types.Address `json:"to"`
	Amount      *big.Int      `json:"amount"`
	Power       int64         `json:"power"`
	BlockReward *big.Int      `json:"blockReward"`
	Reward      *big.Int      `json:"receivedReward"`

	TxHash       types.HexBytes `json:"txhash"`
	StartHeight  int64          `json:"startHeight"`
	RefundHeight int64          `json:"refundHeight"`

	mtx sync.RWMutex
}

func NewStakeWithAmount(from, to types.Address, amt *big.Int, height int64, txhash types.HexBytes, rules types.IGovRuleHandler) *Stake {
	power := rules.AmountToPower(amt)
	blockReward := rules.PowerToReward(power)
	return &Stake{
		From:         from,
		To:           to,
		Amount:       amt,
		Power:        power,
		BlockReward:  blockReward,
		Reward:       big.NewInt(0),
		StartHeight:  height,
		RefundHeight: 0,
		TxHash:       txhash,
	}
}

func NewStakeWithPower(owner, to types.Address, power int64, height int64, txhash types.HexBytes, rules types.IGovRuleHandler) *Stake {
	amt := rules.PowerToAmount(power)
	blockReward := rules.PowerToReward(power)
	return &Stake{
		From:         owner,
		To:           to,
		Amount:       amt,
		Power:        power,
		BlockReward:  blockReward,
		Reward:       big.NewInt(0),
		StartHeight:  height,
		RefundHeight: 0,
		TxHash:       txhash,
	}
}

func (s *Stake) Equal(o *Stake) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return bytes.Compare(s.From, o.From) == 0 &&
		bytes.Compare(s.To, o.To) == 0 &&
		bytes.Compare(s.TxHash, o.TxHash) == 0 &&
		s.StartHeight == o.StartHeight &&
		s.Amount.Cmp(o.Amount) == 0
}

func (s *Stake) Copy() *Stake {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return &Stake{
		From:         append(s.From, nil...),
		To:           append(s.To, nil...),
		Amount:       new(big.Int).Set(s.Amount),
		Power:        s.Power,
		BlockReward:  new(big.Int).Set(s.BlockReward),
		Reward:       new(big.Int).Set(s.Reward),
		StartHeight:  s.StartHeight,
		RefundHeight: s.RefundHeight,
	}
}

func (s *Stake) ApplyReward() *big.Int {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.applyReward()
}

func (s *Stake) applyReward() *big.Int {
	s.Reward = new(big.Int).Add(s.Reward, s.BlockReward)
	return s.BlockReward
}

func (s *Stake) IsSelfStake() bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return bytes.Compare(s.From, s.To) == 0
}

func (s *Stake) String() string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	bz, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("{error: %v}", err)
	}
	return string(bz)
}
