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

var (
	amountPerPower = uint256.NewInt(1_000000000_000000000) // 1RIGO == 1Power
)

type GovRule struct {
	version               int64
	maxValidatorCnt       int64
	minValidatorStake     *uint256.Int
	rewardPerPower        int64
	lazyRewardBlocks      int64
	lazyApplyingBlocks    int64
	minTrxFee             *uint256.Int
	minVotingPeriodBlocks int64
	maxVotingPeriodBlocks int64

	minSelfStakeRatio      int64
	maxUpdatableStakeRatio int64 // todo: add max updatable stake: issue #34
	slashRatio             int64

	mtx sync.RWMutex
}

func DefaultGovRule() *GovRule {
	return &GovRule{
		version:           1,
		maxValidatorCnt:   21,
		minValidatorStake: uint256.MustFromDecimal("7000000000000000000000000"), // 7,000,000 RIGO
		// block interval = 6s
		// blocks/1Y = 5_256_000
		// issuance rate : 5%  => 0.05RIGO / 1RIGO(Power),1Y(5_256_000 blocks)
		// block reward = 9_512_937_595.1293759513 / 1RIGO & 1 block
		rewardPerPower:         9_512_937_595,
		lazyRewardBlocks:       2592000, // = 60 * 60 * 24 * 30 => 30 days
		lazyApplyingBlocks:     259200,  // = 60 * 60 * 24 * 3 => 3 days
		minTrxFee:              uint256.NewInt(10),
		minVotingPeriodBlocks:  259200,  // = 60 * 60 * 24 * 3 => 3 days
		maxVotingPeriodBlocks:  2592000, // = 60 * 60 * 24 * 30 => 30 days
		minSelfStakeRatio:      50,      // 50%
		maxUpdatableStakeRatio: 30,      // 30%
		slashRatio:             50,      // 50%
	}
}

// 9512937595

func Test1GovRule() *GovRule {
	return &GovRule{
		version:                1,
		maxValidatorCnt:        10,
		minValidatorStake:      uint256.MustFromDecimal("10000000000000000000"), // 10 RIGO
		rewardPerPower:         2_000_000_000,
		lazyRewardBlocks:       10,
		lazyApplyingBlocks:     10,
		minTrxFee:              uint256.NewInt(20),
		minVotingPeriodBlocks:  10,
		maxVotingPeriodBlocks:  10,
		minSelfStakeRatio:      50, // 50%
		maxUpdatableStakeRatio: 30, // 30%
		slashRatio:             50, // 50%
	}
}

func Test2GovRule() *GovRule {
	return &GovRule{
		version:                2,
		maxValidatorCnt:        10,
		minValidatorStake:      uint256.MustFromDecimal("5000000000000000000"), // 5 RIGO
		rewardPerPower:         2_000_000_000,
		lazyRewardBlocks:       30,
		lazyApplyingBlocks:     40,
		minTrxFee:              uint256.NewInt(20),
		minVotingPeriodBlocks:  50,
		maxVotingPeriodBlocks:  60,
		minSelfStakeRatio:      50, // 50%
		maxUpdatableStakeRatio: 30, // 30%
		slashRatio:             50, // 50%
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
	r.minValidatorStake = new(uint256.Int).SetBytes(pm.XMinValidatorStake)
	r.rewardPerPower = pm.RewardPerPower
	r.lazyRewardBlocks = pm.LazyRewardBlocks
	r.lazyApplyingBlocks = pm.LazyApplyingBlocks
	r.minTrxFee = new(uint256.Int).SetBytes(pm.XMinTrxFee)
	r.minVotingPeriodBlocks = pm.MinVotingPeriodBlocks
	r.maxVotingPeriodBlocks = pm.MaxVotingPeriodBlocks
	r.minSelfStakeRatio = pm.MinSelfStakeRatio
	r.maxUpdatableStakeRatio = pm.MaxUpdatableStakeRatio
	r.slashRatio = pm.SlashRatio
}

func (r *GovRule) toProto() *GovRuleProto {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	a := &GovRuleProto{
		Version:                r.version,
		MaxValidatorCnt:        r.maxValidatorCnt,
		XMinValidatorStake:     r.minValidatorStake.Bytes(),
		RewardPerPower:         r.rewardPerPower,
		LazyRewardBlocks:       r.lazyRewardBlocks,
		LazyApplyingBlocks:     r.lazyApplyingBlocks,
		XMinTrxFee:             r.minTrxFee.Bytes(),
		MinVotingPeriodBlocks:  r.minVotingPeriodBlocks,
		MaxVotingPeriodBlocks:  r.maxVotingPeriodBlocks,
		MinSelfStakeRatio:      r.minSelfStakeRatio,
		MaxUpdatableStakeRatio: r.maxUpdatableStakeRatio,
		SlashRatio:             r.slashRatio,
	}
	return a
}

func (r *GovRule) MarshalJSON() ([]byte, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	tm := &struct {
		Version                int64  `json:"version"`
		MaxValidatorCnt        int64  `json:"maxValidatorCnt"`
		MinValidatorStake      string `json:"minValidatorStake"`
		RewardPerPower         int64  `json:"rewardPerPower"`
		LazyRewardBlocks       int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks     int64  `json:"lazyApplyingBlocks"`
		MinTrxFee              string `json:"minTrxFee"`
		MinVotingBlocks        int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks        int64  `json:"maxVotingPeriodBlocks"`
		MinSelfStakeRatio      int64  `json:"minSelfStakeRatio"`
		MaxUpdatableStakeRatio int64  `json:"maxUpdatableStakeRatio"`
		SlashRatio             int64  `json:"slashRatio"`
	}{
		Version:                r.version,
		MaxValidatorCnt:        r.maxValidatorCnt,
		MinValidatorStake:      r.minValidatorStake.String(), // hex-string
		RewardPerPower:         r.rewardPerPower,             // hex-string
		LazyRewardBlocks:       r.lazyRewardBlocks,
		LazyApplyingBlocks:     r.lazyApplyingBlocks,
		MinTrxFee:              r.minTrxFee.String(),
		MinVotingBlocks:        r.minVotingPeriodBlocks,
		MaxVotingBlocks:        r.maxVotingPeriodBlocks,
		MinSelfStakeRatio:      r.minSelfStakeRatio,
		MaxUpdatableStakeRatio: r.maxUpdatableStakeRatio,
		SlashRatio:             r.slashRatio,
	}
	return tmjson.Marshal(tm)
}

func (r *GovRule) UnmarshalJSON(bz []byte) error {
	tm := &struct {
		Version                int64  `json:"version"`
		MaxValidatorCnt        int64  `json:"maxValidatorCnt"`
		MinValidatorStake      string `json:"minValidatorStake"`
		RewardPerPower         int64  `json:"rewardPerPower"`
		LazyRewardBlocks       int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks     int64  `json:"lazyApplyingBlocks"`
		MinTrxFee              string `json:"minTrxFee"`
		MinVotingBlocks        int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks        int64  `json:"maxVotingPeriodBlocks"`
		MinSelfStakeRatio      int64  `json:"minSelfStakeRatio"`
		MaxUpdatableStakeRatio int64  `json:"maxUpdatableStakeRatio"`
		SlashRatio             int64  `json:"slashRatio"`
	}{}

	err := tmjson.Unmarshal(bz, tm)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.version = tm.Version
	r.maxValidatorCnt = tm.MaxValidatorCnt
	r.minValidatorStake, err = uint256.FromHex(tm.MinValidatorStake)
	if err != nil {
		return err
	}
	r.rewardPerPower = tm.RewardPerPower
	r.lazyRewardBlocks = tm.LazyRewardBlocks
	r.lazyApplyingBlocks = tm.LazyApplyingBlocks
	r.minTrxFee, err = uint256.FromHex(tm.MinTrxFee)
	if err != nil {
		return err
	}
	r.minVotingPeriodBlocks = tm.MinVotingBlocks
	r.maxVotingPeriodBlocks = tm.MaxVotingBlocks
	r.minSelfStakeRatio = tm.MinSelfStakeRatio
	r.maxUpdatableStakeRatio = tm.MaxUpdatableStakeRatio
	r.slashRatio = tm.SlashRatio
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

func (r *GovRule) MinValidatorStake() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.minValidatorStake)
}

func (r *GovRule) RewardPerPower() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.rewardPerPower
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
func (r *GovRule) MinSelfStakeRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minSelfStakeRatio
}
func (r *GovRule) MaxUpdatableStakeRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxUpdatableStakeRatio
}
func (r *GovRule) SlashRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.slashRatio
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

//
// utility methods

// MaxStakeAmount means the max of amount which could be deposited.
// tmtypes.MaxTotalVotingPower = int64(math.MaxInt64) / 8
// When the type of voting power is `int64`and VP:XCO = 1:1,
// the MAXSTAKEsau becomes `int64(math.MaxInt64) / 8 * 10^18` (~= 922경 XCO)
func MaxStakeAmount() *uint256.Int {
	return new(uint256.Int).Mul(uint256.NewInt(uint64(tmtypes.MaxTotalVotingPower)), amountPerPower)
}

func MaxTotalPower() int64 {
	return tmtypes.MaxTotalVotingPower
}

func AmountToPower(amt *uint256.Int) int64 {
	// 1 VotingPower == 1 RIGO
	_vp := new(uint256.Int).Div(amt, amountPerPower)
	vp := int64(_vp.Uint64())
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func PowerToAmount(power int64) *uint256.Int {
	// 1 VotingPower == 1 RIGO
	return new(uint256.Int).Mul(uint256.NewInt(uint64(power)), amountPerPower)
}

func AmountPerPower() *uint256.Int {
	return amountPerPower.Clone()
}

var _ ledger.ILedgerItem = (*GovRule)(nil)
var _ IGovHandler = (*GovRule)(nil)
