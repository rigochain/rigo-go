package stake

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
	"sync"
)

type FeeReward struct {
	pubKey      types.HexBytes
	receivedFee *big.Int

	mtx sync.RWMutex
}

func NewFeeReward(pubKey types.HexBytes) *FeeReward {
	return &FeeReward{
		pubKey:      pubKey,
		receivedFee: big.NewInt(0),
	}
}

func (fr *FeeReward) Add(amt *big.Int) error {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	}

	_ = fr.receivedFee.Add(fr.receivedFee, amt)
	return nil
}

func (fr *FeeReward) Sub(amt *big.Int) error {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()

	if amt.Sign() < 0 {
		return xerrors.ErrNegAmount
	} else if amt.Cmp(fr.receivedFee) > 0 {
		return xerrors.New("too big number")
	}

	_ = fr.receivedFee.Add(fr.receivedFee, amt)
	return nil
}

func (fr *FeeReward) PubKey() types.HexBytes {
	fr.mtx.RLock()
	defer fr.mtx.RUnlock()

	return fr.pubKey
}

func (fr *FeeReward) ReceivedFee() *big.Int {
	fr.mtx.RLock()
	defer fr.mtx.RUnlock()

	return new(big.Int).Set(fr.receivedFee)
}

func (fr *FeeReward) MarshalJSON() ([]byte, error) {
	_tmp := &struct {
		PubKey      types.HexBytes `json:"pubKey"`
		ReceivedFee string         `json:"receivedFee""`
	}{
		PubKey:      fr.pubKey,
		ReceivedFee: fr.receivedFee.String(),
	}
	return json.Marshal(_tmp)
}

func (fr *FeeReward) UnmarshalJSON(bz []byte) error {
	_tmp := &struct {
		PubKey      types.HexBytes `json:"pubKey"`
		ReceivedFee string         `json:"receivedFee""`
	}{}

	if err := json.Unmarshal(bz, _tmp); err != nil {
		return err
	}

	fr.pubKey = _tmp.PubKey
	if _, ok := fr.receivedFee.SetString(_tmp.ReceivedFee, 10); !ok {
		return xerrors.New("wrong received fee amount: " + _tmp.ReceivedFee)
	}
	return nil
}

func (fr *FeeReward) Encode() ([]byte, error) {
	return json.Marshal(fr)
}

func (fr *FeeReward) Decode(bz []byte) error {
	return json.Unmarshal(bz, fr)
}
