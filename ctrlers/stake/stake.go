package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"math/big"
	"sort"
	"sync"
)

type Stake struct {
	From            types.Address `json:"owner"`
	To              types.Address `json:"to"`
	Amount          *big.Int      `json:"amount"`
	Power           int64         `json:"power"`
	BlockRewardUnit *big.Int      `json:"blockRewardUnit"`
	ReceivedReward  *big.Int      `json:"ReceivedReward"`

	TxHash       abytes.HexBytes `json:"txhash"`
	StartHeight  int64           `json:"startHeight"`
	RefundHeight int64           `json:"refundHeight"`

	mtx sync.RWMutex
}

func (s *Stake) Key() ledger.LedgerKey {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return ledger.ToLedgerKey(s.TxHash)
}

func (s *Stake) Encode() ([]byte, xerrors.XError) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if bz, err := json.Marshal(s); err != nil {
		return nil, xerrors.NewFrom(err)
	} else {
		return bz, nil
	}
}

func (s *Stake) Decode(d []byte) xerrors.XError {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if err := json.Unmarshal(d, s); err != nil {
		return xerrors.NewFrom(err)
	}
	return nil
}

var _ ledger.ILedgerItem = (*Stake)(nil)

func NewStakeWithAmount(from, to types.Address, amt *big.Int, height int64, txhash abytes.HexBytes, govHelper ctrlertypes.IGovHelper) *Stake {
	power := govHelper.AmountToPower(amt)
	blockReward := govHelper.PowerToReward(power)
	return &Stake{
		From:            from,
		To:              to,
		Amount:          amt,
		Power:           power,
		BlockRewardUnit: blockReward,
		ReceivedReward:  big.NewInt(0),
		StartHeight:     height,
		RefundHeight:    0,
		TxHash:          txhash,
	}
}

func NewStakeWithPower(owner, to types.Address, power int64, height int64, txhash abytes.HexBytes, govHelper ctrlertypes.IGovHelper) *Stake {
	amt := govHelper.PowerToAmount(power)
	blockReward := govHelper.PowerToReward(power)
	return &Stake{
		From:            owner,
		To:              to,
		Amount:          amt,
		Power:           power,
		BlockRewardUnit: blockReward,
		ReceivedReward:  big.NewInt(0),
		StartHeight:     height,
		RefundHeight:    0,
		TxHash:          txhash,
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
		From:            append(s.From, nil...),
		To:              append(s.To, nil...),
		Amount:          new(big.Int).Set(s.Amount),
		Power:           s.Power,
		BlockRewardUnit: new(big.Int).Set(s.BlockRewardUnit),
		ReceivedReward:  new(big.Int).Set(s.ReceivedReward),
		StartHeight:     s.StartHeight,
		RefundHeight:    s.RefundHeight,
	}
}

func (s *Stake) ApplyReward() *big.Int {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.applyReward()
}

func (s *Stake) applyReward() *big.Int {
	s.ReceivedReward = new(big.Int).Add(s.ReceivedReward, s.BlockRewardUnit)
	return s.BlockRewardUnit
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
