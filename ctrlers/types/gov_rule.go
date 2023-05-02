package types

import (
	"encoding/json"
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/protobuf/proto"
	"sync"
)

type GovRule struct {
	version               int64
	maxValidatorCnt       int64
	amountPerPower        *uint256.Int
	rewardPerPower        *uint256.Int
	lazyRewardBlocks      int64
	lazyApplyingBlocks    int64
	minTrxFee             *uint256.Int
	minVotingPeriodBlocks int64
	maxVotingPeriodBlocks int64

	minSelfStakeRatio      int64 // todo: add min equity stake
	maxUpdatableStakeRatio int64 // todo: add max updatable stake

	mtx sync.RWMutex
}

func DefaultGovRule() *GovRule {
	return &GovRule{
		version:               0,
		maxValidatorCnt:       21,
		amountPerPower:        uint256.NewInt(1_000000000_000000000),
		rewardPerPower:        uint256.NewInt(1_000000000),
		lazyRewardBlocks:      20,
		lazyApplyingBlocks:    10,
		minTrxFee:             uint256.NewInt(10),
		minVotingPeriodBlocks: 259200,  // = 60 * 60 * 24 * 3, // 3 days
		maxVotingPeriodBlocks: 2592000, // = 60 * 60 * 24 * 30,    // 30 days
	}
}

func Test1GovRule() *GovRule {
	return &GovRule{
		version:               1,
		maxValidatorCnt:       10,
		amountPerPower:        uint256.NewInt(1_000000000),
		rewardPerPower:        uint256.NewInt(2_000000000),
		lazyRewardBlocks:      30,
		lazyApplyingBlocks:    40,
		minTrxFee:             uint256.NewInt(20),
		minVotingPeriodBlocks: 50,
		maxVotingPeriodBlocks: 60,
	}
}

func Test2GovRule() *GovRule {
	return &GovRule{
		version:               2,
		maxValidatorCnt:       10,
		amountPerPower:        uint256.NewInt(1_000000000),
		rewardPerPower:        uint256.NewInt(2_000000000),
		lazyRewardBlocks:      30,
		lazyApplyingBlocks:    40,
		minTrxFee:             uint256.NewInt(20),
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
		return xerrors.From(err)
	}
	r.fromProto(pm)
	return nil
}

func (r *GovRule) Encode() ([]byte, xerrors.XError) {
	if bz, err := proto.Marshal(r.toProto()); err != nil {
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (r *GovRule) fromProto(pm *GovRuleProto) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.version = pm.Version
	r.maxValidatorCnt = pm.MaxValidatorCnt
	r.amountPerPower = new(uint256.Int).SetBytes(pm.XAmountPerPower)
	r.rewardPerPower = new(uint256.Int).SetBytes(pm.XRewardPerPower)
	r.lazyRewardBlocks = pm.LazyRewardBlocks
	r.lazyApplyingBlocks = pm.LazyApplyingBlocks
	r.minTrxFee = new(uint256.Int).SetBytes(pm.XMinTrxFee)
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
		AmountPerPower:     r.amountPerPower.String(), // hex-string
		RewardPerPower:     r.rewardPerPower.String(), // hex-string
		LazyRewardBlocks:   r.lazyRewardBlocks,
		LazyApplyingBlocks: r.lazyApplyingBlocks,
		MinTrxFee:          r.minTrxFee.String(),
		MinVotingBlocks:    r.minVotingPeriodBlocks,
		MaxVotingBlocks:    r.maxVotingPeriodBlocks,
	}
	return tmjson.Marshal(tm)
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

	if err := tmjson.Unmarshal(bz, tm); err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.version = tm.Version
	r.maxValidatorCnt = tm.MaxValidatorCnt
	r.amountPerPower = uint256.MustFromHex(tm.AmountPerPower)
	r.rewardPerPower = uint256.MustFromHex(tm.RewardPerPower)
	r.lazyRewardBlocks = tm.LazyRewardBlocks
	r.lazyApplyingBlocks = tm.LazyApplyingBlocks
	r.minTrxFee = uint256.MustFromHex(tm.MinTrxFee)
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

func (r *GovRule) AmountPerPower() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.amountPerPower)
}

func (r *GovRule) RewardPerPower() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.rewardPerPower)
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

func (r *GovRule) MinTrxFee() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.minTrxFee)
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
func (r *GovRule) MaxStakeAmount() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Mul(uint256.NewInt(uint64(tmtypes.MaxTotalVotingPower)), r.amountPerPower)
}

func (r *GovRule) MaxTotalPower() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return tmtypes.MaxTotalVotingPower
}

func (r *GovRule) AmountToPower(amt *uint256.Int) int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	_vp := new(uint256.Int).Div(amt, r.amountPerPower)
	vp := int64(_vp.Uint64())
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (r *GovRule) PowerToAmount(power int64) *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	// 1 VotingPower == 1 XCO
	return new(uint256.Int).Mul(uint256.NewInt(uint64(power)), r.amountPerPower)
}

func (r *GovRule) PowerToReward(power int64) *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(uint256.Int).Mul(uint256.NewInt(uint64(power)), r.rewardPerPower)
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
