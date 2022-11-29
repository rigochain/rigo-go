package stake

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
	"sync"
)

type Reward struct {
	block *big.Int   `json:"block"`
	fee   *FeeReward `json:"fee"`

	mtx sync.RWMutex
}

func NewReward(pubKey types.HexBytes) *Reward {
	return &Reward{
		block: big.NewInt(0),
		fee:   NewFeeReward(pubKey),
	}
}

func (r *Reward) BlockReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.block)
}

func (r *Reward) FeeReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.fee.ReceivedFee()
}

func (r *Reward) TotalReward() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Add(r.block, r.fee.receivedFee)
}

func (r *Reward) AddBlockReward(amt *big.Int) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}

	_ = r.block.Add(r.block, amt)
	return nil
}

func (r *Reward) AddFeeReward(amt *big.Int) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.fee.Add(amt)
}

func (r *Reward) SubBlockReward(amt *big.Int) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.block.Cmp(amt) < 0 {
		return xerrors.New("too big amount")
	}
	_ = r.block.Sub(r.block, amt)
	return nil
}

func (r *Reward) SubFeeReward(amt *big.Int) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.fee.Sub(amt)
}

func (r *Reward) MarshalJSON() ([]byte, error) {
	_tmp := &struct {
		Block *big.Int   `json:"block"`
		Fee   *FeeReward `json:"fee"`
	}{
		Block: r.block,
		Fee:   r.fee,
	}
	return json.Marshal(_tmp)
}

func (r *Reward) UnmarshalJSON(bz []byte) error {
	_tmp := &struct {
		Block *big.Int   `json:"block"`
		Fee   *FeeReward `json:"fee"`
	}{}

	if err := json.Unmarshal(bz, _tmp); err != nil {
		return err
	}

	r.block = _tmp.Block
	r.fee = _tmp.Fee
	return nil
}
