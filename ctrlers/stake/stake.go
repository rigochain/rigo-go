package stake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"sort"
	"sync"
)

type Stake struct {
	From            types.Address `json:"owner"`
	To              types.Address `json:"to"`
	Amount          *uint256.Int  `json:"amount"`
	Power           int64         `json:"power,string"`
	BlockRewardUnit *uint256.Int  `json:"blockRewardUnit"`
	ReceivedReward  *uint256.Int  `json:"ReceivedReward"`

	TxHash       abytes.HexBytes `json:"txhash"`
	StartHeight  int64           `json:"startHeight,string"`
	RefundHeight int64           `json:"refundHeight,string"`

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
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (s *Stake) Decode(d []byte) xerrors.XError {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if err := json.Unmarshal(d, s); err != nil {
		return xerrors.From(err)
	}
	return nil
}

var _ ledger.ILedgerItem = (*Stake)(nil)

func NewStakeWithAmount(from, to types.Address, amt *uint256.Int, height int64, txhash abytes.HexBytes, govHandler ctrlertypes.IGovHandler) *Stake {
	power := govHandler.AmountToPower(amt)
	blockReward := govHandler.PowerToReward(power)
	return &Stake{
		From:            from,
		To:              to,
		Amount:          amt,
		Power:           power,
		BlockRewardUnit: blockReward,
		ReceivedReward:  uint256.NewInt(0),
		StartHeight:     height,
		RefundHeight:    0,
		TxHash:          txhash,
	}
}

func NewStakeWithPower(owner, to types.Address, power int64, height int64, txhash abytes.HexBytes, govHandler ctrlertypes.IGovHandler) *Stake {
	amt := govHandler.PowerToAmount(power)
	blockReward := govHandler.PowerToReward(power)
	return &Stake{
		From:            owner,
		To:              to,
		Amount:          amt,
		Power:           power,
		BlockRewardUnit: blockReward,
		ReceivedReward:  uint256.NewInt(0),
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
		Amount:          new(uint256.Int).Set(s.Amount),
		Power:           s.Power,
		BlockRewardUnit: new(uint256.Int).Set(s.BlockRewardUnit),
		ReceivedReward:  new(uint256.Int).Set(s.ReceivedReward),
		StartHeight:     s.StartHeight,
		RefundHeight:    s.RefundHeight,
	}
}

func (s *Stake) ApplyReward() *uint256.Int {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.applyReward()
}

func (s *Stake) applyReward() *uint256.Int {
	_ = s.ReceivedReward.Add(s.ReceivedReward, s.BlockRewardUnit)
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
