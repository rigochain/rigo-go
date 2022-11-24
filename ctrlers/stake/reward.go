package stake

import (
	"math/big"
	"sync"
)

type Reward struct {
	BlockReward *big.Int `json:"blockReward"`
	FeeReward   *big.Int `json:"feeReward"`

	mtx sync.RWMutex
}

func NewReward() *Reward {
	return &Reward{
		BlockReward: big.NewInt(0),
		FeeReward:   big.NewInt(0),
	}
}

func (r *Reward) GetBlockReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.BlockReward)
}

func (r *Reward) GetFeeReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.FeeReward)
}

func (r *Reward) GetTotalReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Add(r.BlockReward, r.FeeReward)
}

func (r *Reward) AddBlockReward(amt *big.Int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	_ = r.BlockReward.Add(r.BlockReward, amt)
}

func (r *Reward) AddFeeReward(amt *big.Int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	_ = r.FeeReward.Add(r.FeeReward, amt)
}

func (r *Reward) SubBlockReward(amt *big.Int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.BlockReward.Cmp(amt) >= 0 {
		_ = r.BlockReward.Sub(r.BlockReward, amt)
	}
}

func (r *Reward) SubFeeReward(amt *big.Int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.FeeReward.Cmp(amt) >= 0 {
		_ = r.FeeReward.Sub(r.FeeReward, amt)
	}
}
