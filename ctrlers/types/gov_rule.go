package types

import (
	"fmt"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/protobuf/proto"
	"math/big"
	"sync"
)

type GovRule struct {
	version               int64
	maxValidatorCnt       int64
	amountPerPower        *big.Int
	rewardPerPower        *big.Int
	lazyRewardBlocks      int64
	lazyApplyingBlocks    int64
	minTrxFee             *big.Int
	minVotingPeriodBlocks int64
	maxVotingPeriodBlocks int64

	mtx sync.RWMutex
}

func DefaultGovRule() *GovRule {
	return &GovRule{
		version:               0,
		maxValidatorCnt:       21,
		amountPerPower:        big.NewInt(1_000000000_000000000),
		rewardPerPower:        big.NewInt(1_000000000),
		lazyRewardBlocks:      20,
		lazyApplyingBlocks:    10,
		minTrxFee:             big.NewInt(10),
		minVotingPeriodBlocks: 259200,  // = 60 * 60 * 24 * 3, // 3 days
		maxVotingPeriodBlocks: 2592000, // = 60 * 60 * 24 * 30,    // 30 days
	}
}

func Test1GovRule() *GovRule {
	return &GovRule{
		version:               1,
		maxValidatorCnt:       10,
		amountPerPower:        big.NewInt(1_000000000),
		rewardPerPower:        big.NewInt(2_000000000),
		lazyRewardBlocks:      30,
		lazyApplyingBlocks:    40,
		minTrxFee:             big.NewInt(20),
		minVotingPeriodBlocks: 50,
		maxVotingPeriodBlocks: 60,
	}
}

func Test2GovRule() *GovRule {
	return &GovRule{
		version:               2,
		maxValidatorCnt:       10,
		amountPerPower:        big.NewInt(1_000000000),
		rewardPerPower:        big.NewInt(2_000000000),
		lazyRewardBlocks:      30,
		lazyApplyingBlocks:    40,
		minTrxFee:             big.NewInt(20),
		minVotingPeriodBlocks: 50,
		maxVotingPeriodBlocks: 60,
	}
}

func DecodeGovRule(bz []byte) (*GovRule, xerrors.XError) {
	ret := &GovRule{}
	if xerr := ret.Decode(bz); xerr != nil {
		return nil, xerr
	}
	return ret, nil
}

func (r *GovRule) Key() ledger.LedgerKey {
	return ledger.ToLedgerKey(bytes.ZeroBytes(32))
}

func (r *GovRule) Decode(bz []byte) xerrors.XError {
	pm := &GovRuleProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.NewFrom(err)
	}
	r.fromProto(pm)
	return nil
}

func (r *GovRule) Encode() ([]byte, xerrors.XError) {
	if bz, err := proto.Marshal(r.toProto()); err != nil {
		return nil, xerrors.NewFrom(err)
	} else {
		return bz, nil
	}
}

func (r *GovRule) fromProto(pm *GovRuleProto) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.version = pm.Version
	r.maxValidatorCnt = pm.MaxValidatorCnt
	r.amountPerPower = new(big.Int).SetBytes(pm.XAmountPerPower)
	r.rewardPerPower = new(big.Int).SetBytes(pm.XRewardPerPower)
	r.lazyRewardBlocks = pm.LazyRewardBlocks
	r.lazyApplyingBlocks = pm.LazyApplyingBlocks
	r.minTrxFee = new(big.Int).SetBytes(pm.XMinTrxFee)
	r.minVotingPeriodBlocks = pm.MinVotingPeriodBlocks
	r.maxVotingPeriodBlocks = pm.MaxVotingPeriodBlocks
}

func (r *GovRule) toProto() *GovRuleProto {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	a := &GovRuleProto{
		Version:               r.version,
		MaxValidatorCnt:       r.maxValidatorCnt,
		XAmountPerPower:       r.amountPerPower.Bytes(),
		XRewardPerPower:       r.rewardPerPower.Bytes(),
		LazyRewardBlocks:      r.lazyRewardBlocks,
		LazyApplyingBlocks:    r.lazyApplyingBlocks,
		XMinTrxFee:            r.minTrxFee.Bytes(),
		MinVotingPeriodBlocks: r.minVotingPeriodBlocks,
		MaxVotingPeriodBlocks: r.maxVotingPeriodBlocks,
	}
	return a
}

func (r *GovRule) MarshalJSON() ([]byte, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	tm := &struct {
		Version            int64  `json:"version"`
		MaxValidatorCnt    int64  `json:"maxValidatorCnt"`
		AmountPerPower     string `json:"amountPerPower"`
		RewardPerPower     string `json:"rewardPerPower"`
		LazyRewardBlocks   int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks int64  `json:"lazyApplyingBlocks"`
		MinTrxFee          string `json:"minTrxFee"`
		MinVotingBlocks    int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks    int64  `json:"maxVotingPeriodBlocks"`
	}{
		Version:            r.version,
		MaxValidatorCnt:    r.maxValidatorCnt,
		AmountPerPower:     r.amountPerPower.String(),
		RewardPerPower:     r.rewardPerPower.String(),
		LazyRewardBlocks:   r.lazyRewardBlocks,
		LazyApplyingBlocks: r.lazyApplyingBlocks,
		MinTrxFee:          r.minTrxFee.String(),
		MinVotingBlocks:    r.minVotingPeriodBlocks,
		MaxVotingBlocks:    r.maxVotingPeriodBlocks,
	}
	return json.Marshal(tm)
}

func (r *GovRule) UnmarshalJSON(bz []byte) error {
	tm := &struct {
		Version            int64  `json:"version"`
		MaxValidatorCnt    int64  `json:"maxValidatorCnt"`
		AmountPerPower     string `json:"amountPerPower"`
		RewardPerPower     string `json:"rewardPerPower"`
		LazyRewardBlocks   int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks int64  `json:"lazyApplyingBlocks"`
		MinTrxFee          string `json:"minTrxFee"`
		MinVotingBlocks    int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks    int64  `json:"maxVotingPeriodBlocks"`
	}{}

	if err := json.Unmarshal(bz, tm); err != nil {
		return err
	}

	amtPower, ok := new(big.Int).SetString(tm.AmountPerPower, 10)
	if !ok {
		return xerrors.New("amountPerPower is wrong")
	}
	rwdPower, ok := new(big.Int).SetString(tm.RewardPerPower, 10)
	if !ok {
		return xerrors.New("rewardPerPower is wrong")
	}
	minFee, ok := new(big.Int).SetString(tm.MinTrxFee, 10)
	if !ok {
		return xerrors.New("minTrxFee is wrong")
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.version = tm.Version
	r.maxValidatorCnt = tm.MaxValidatorCnt
	r.amountPerPower = amtPower
	r.rewardPerPower = rwdPower
	r.lazyRewardBlocks = tm.LazyRewardBlocks
	r.lazyApplyingBlocks = tm.LazyApplyingBlocks
	r.minTrxFee = minFee
	r.minVotingPeriodBlocks = tm.MinVotingBlocks
	r.maxVotingPeriodBlocks = tm.MaxVotingBlocks
	return nil
}

func (r *GovRule) Version() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.version
}

func (r *GovRule) MaxValidatorCnt() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxValidatorCnt
}

func (r *GovRule) AmountPerPower() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.amountPerPower)
}

func (r *GovRule) RewardPerPower() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.rewardPerPower)
}

func (r *GovRule) LazyRewardBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.lazyRewardBlocks
}

func (r *GovRule) LazyApplyingBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.lazyApplyingBlocks
}

func (r *GovRule) MinTrxFee() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Set(r.minTrxFee)
}

func (r *GovRule) MinVotingPeriodBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minVotingPeriodBlocks
}

func (r *GovRule) MaxVotingPeriodBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxVotingPeriodBlocks
}

//
// utility methods

// MaxStakeAmount means the max of amount which could be deposited.
// tmtypes.MaxTotalVotingPower = int64(math.MaxInt64) / 8
// When the type of voting power is `int64`and VP:XCO = 1:1,
// the MAXSTAKEsau becomes `int64(math.MaxInt64) / 8 * 10^18` (~= 922경 XCO)
func (r *GovRule) MaxStakeAmount() *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(big.Int).Mul(big.NewInt(tmtypes.MaxTotalVotingPower), r.amountPerPower)
}

func (r *GovRule) MaxTotalPower() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return tmtypes.MaxTotalVotingPower
}

func (r *GovRule) AmountToPower(amt *big.Int) int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, r.amountPerPower)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (r *GovRule) PowerToAmount(power int64) *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), r.amountPerPower)
}

func (r *GovRule) PowerToReward(power int64) *big.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), r.rewardPerPower)
}

func (r *GovRule) String() string {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	if bz, err := json.MarshalIndent(r, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(bz)
	}
}

var _ ledger.ILedgerItem = (*GovRule)(nil)
var _ IGovHelper = (*GovRule)(nil)
