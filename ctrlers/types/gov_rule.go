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
	"math"
	"sync"
)

var (
	amountPerPower = uint256.NewInt(1_000000000_000000000) // 1RIGO == 1Power
)

type GovRule struct {
	version               int64
	maxValidatorCnt       int64
	minValidatorStake     *uint256.Int
	rewardPerPower        *uint256.Int
	lazyRewardBlocks      int64
	lazyApplyingBlocks    int64
	gasPrice              *uint256.Int
	minTrxGas             uint64
	maxTrxGas             uint64
	maxBlockGas           uint64
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
		//
		// issue #60
		//
		// block interval = 3s
		// min blocks/1Y = 10,512,000 (if all blocks interval 3s)
		// max blocks/1Y = 31,536,000 (if all blocks interval 1s)
		// 1RIGO = 1POWer = 10^18 amount
		//
		// When the min issuance rate is 5%,
		// 			= 0.05 RIGO [per 1Power(RIGO),1Y(10_512_000 blocks)]
		//			= 0.05 RIGO / 10_512_000 blocks [per 1Power(RIGO), 1block]
		//          = 50,000,000,000,000,000 / 10,512,000 [per 1Power(RIGO), 1block]
		//			= 4,756,468,797.5646879756 [per 1Power(RIGO), 1block]
		// the `rewardPerPower` should be 4,756,468,797.5646879756.
		// When like this,
		// the max issuance rate becomes ...
		//			= 4,756,468,797 * 31,536,000(blocks in 1Y)
		//			= 149,999,999,982,192,000 amount [per 1RIGO, 1Y]
		// , it's about 15% of 1 power
		rewardPerPower:         uint256.NewInt(4_756_468_797),  // amount
		lazyRewardBlocks:       2592000,                        // = 60 * 60 * 24 * 30 => 30 days
		lazyApplyingBlocks:     259200,                         // = 60 * 60 * 24 * 3 => 3 days
		gasPrice:               uint256.NewInt(10_000_000_000), // 10Gwei
		minTrxGas:              uint64(1_000_000_000_000_000),  // 0.001 RIGO = 10^15 wei
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  259200,  // = 60 * 60 * 24 * 3 => 3 days
		maxVotingPeriodBlocks:  2592000, // = 60 * 60 * 24 * 30 => 30 days
		minSelfStakeRatio:      50,      // 50%
		maxUpdatableStakeRatio: 30,      // 30%
		slashRatio:             50,      // 50%
	}
}

func Test1GovRule() *GovRule {
	return &GovRule{
		version:                1,
		maxValidatorCnt:        10,
		minValidatorStake:      uint256.MustFromDecimal("1000000000000000000"), // 1 RIGO
		rewardPerPower:         uint256.NewInt(2_000_000_000),
		lazyRewardBlocks:       10,
		lazyApplyingBlocks:     10,
		gasPrice:               uint256.NewInt(10),
		minTrxGas:              uint64(10),
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
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
		rewardPerPower:         uint256.NewInt(2_000_000_000),
		lazyRewardBlocks:       30,
		lazyApplyingBlocks:     40,
		gasPrice:               uint256.NewInt(20),
		minTrxGas:              uint64(20),
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  50,
		maxVotingPeriodBlocks:  60,
		minSelfStakeRatio:      50, // 50%
		maxUpdatableStakeRatio: 30, // 30%
		slashRatio:             50, // 50%
	}
}

func Test3GovRule() *GovRule {
	return &GovRule{
		version:                4,
		maxValidatorCnt:        13,
		minValidatorStake:      uint256.MustFromDecimal("0"),
		rewardPerPower:         uint256.NewInt(0),
		lazyRewardBlocks:       20,
		lazyApplyingBlocks:     0,
		gasPrice:               nil,
		minTrxGas:              0,
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  0,
		maxVotingPeriodBlocks:  0,
		minSelfStakeRatio:      0,
		maxUpdatableStakeRatio: 10,
		slashRatio:             50,
	}
}

func Test4GovRule() *GovRule {
	return &GovRule{
		version:                4,
		maxValidatorCnt:        13,
		minValidatorStake:      uint256.MustFromDecimal("7000000000000000000000000"),
		rewardPerPower:         uint256.NewInt(4_756_468_797),
		lazyRewardBlocks:       20,
		lazyApplyingBlocks:     259200,
		gasPrice:               uint256.NewInt(10_000_000_000),
		minTrxGas:              uint64(1_000_000_000_000_000),
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  259200,
		maxVotingPeriodBlocks:  2592000,
		minSelfStakeRatio:      50,
		maxUpdatableStakeRatio: 10,
		slashRatio:             50,
	}
}

func Test5GovRule() *GovRule {
	return &GovRule{
		version:                3,
		minValidatorStake:      uint256.MustFromDecimal("0"),
		minSelfStakeRatio:      40,
		maxUpdatableStakeRatio: 50,
		slashRatio:             60,
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
	r.rewardPerPower = new(uint256.Int).SetBytes(pm.XRewardPerPower)
	r.lazyRewardBlocks = pm.LazyRewardBlocks
	r.lazyApplyingBlocks = pm.LazyApplyingBlocks
	r.gasPrice = new(uint256.Int).SetBytes(pm.XGasPrice)
	r.minTrxGas = pm.MinTrxGas
	r.maxTrxGas = pm.MaxTrxGas
	r.maxBlockGas = pm.MaxBlockGas
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
		XRewardPerPower:        r.rewardPerPower.Bytes(),
		LazyRewardBlocks:       r.lazyRewardBlocks,
		LazyApplyingBlocks:     r.lazyApplyingBlocks,
		XGasPrice:              r.gasPrice.Bytes(),
		MinTrxGas:              r.minTrxGas,
		MaxTrxGas:              r.maxTrxGas,
		MaxBlockGas:            r.maxBlockGas,
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
		RewardPerPower         string `json:"rewardPerPower"`
		LazyRewardBlocks       int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks     int64  `json:"lazyApplyingBlocks"`
		GasPrice               string `json:"gasPrice"`
		MinTrxGas              uint64 `json:"minTrxGas"`
		MaxTrxGas              uint64 `json:"maxTrxFee"`
		MaxBlockGas            uint64 `json:"maxBlockFee"`
		MinVotingBlocks        int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks        int64  `json:"maxVotingPeriodBlocks"`
		MinSelfStakeRatio      int64  `json:"minSelfStakeRatio"`
		MaxUpdatableStakeRatio int64  `json:"maxUpdatableStakeRatio"`
		SlashRatio             int64  `json:"slashRatio"`
	}{
		Version:                r.version,
		MaxValidatorCnt:        r.maxValidatorCnt,
		MinValidatorStake:      uint256ToString(r.minValidatorStake), // hex-string
		RewardPerPower:         uint256ToString(r.rewardPerPower),    // hex-string
		LazyRewardBlocks:       r.lazyRewardBlocks,
		LazyApplyingBlocks:     r.lazyApplyingBlocks,
		GasPrice:               uint256ToString(r.gasPrice),
		MinTrxGas:              r.minTrxGas,
		MaxTrxGas:              r.maxTrxGas,
		MaxBlockGas:            r.maxBlockGas,
		MinVotingBlocks:        r.minVotingPeriodBlocks,
		MaxVotingBlocks:        r.maxVotingPeriodBlocks,
		MinSelfStakeRatio:      r.minSelfStakeRatio,
		MaxUpdatableStakeRatio: r.maxUpdatableStakeRatio,
		SlashRatio:             r.slashRatio,
	}
	return tmjson.Marshal(tm)
}

func uint256ToString(value *uint256.Int) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func (r *GovRule) UnmarshalJSON(bz []byte) error {
	tm := &struct {
		Version                int64  `json:"version"`
		MaxValidatorCnt        int64  `json:"maxValidatorCnt"`
		MinValidatorStake      string `json:"minValidatorStake"`
		RewardPerPower         string `json:"rewardPerPower"`
		LazyRewardBlocks       int64  `json:"lazyRewardBlocks"`
		LazyApplyingBlocks     int64  `json:"lazyApplyingBlocks"`
		GasPrice               string `json:"gasPrice"`
		MinTrxGas              uint64 `json:"minTrxGas"`
		MaxTrxGas              uint64 `json:"maxTrxGas"`
		MaxBlockGas            uint64 `json:"maxBlockGas"`
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
	r.minValidatorStake, err = stringToUint256(tm.MinValidatorStake)
	if err != nil {
		return err
	}
	r.rewardPerPower, err = stringToUint256(tm.RewardPerPower)
	if err != nil {
		return err
	}
	r.lazyRewardBlocks = tm.LazyRewardBlocks
	r.lazyApplyingBlocks = tm.LazyApplyingBlocks
	r.gasPrice, err = stringToUint256(tm.GasPrice)
	if err != nil {
		return err
	}
	r.minTrxGas = tm.MinTrxGas
	r.maxTrxGas = tm.MaxTrxGas
	r.maxBlockGas = tm.MaxBlockGas
	r.minVotingPeriodBlocks = tm.MinVotingBlocks
	r.maxVotingPeriodBlocks = tm.MaxVotingBlocks
	r.minSelfStakeRatio = tm.MinSelfStakeRatio
	r.maxUpdatableStakeRatio = tm.MaxUpdatableStakeRatio
	r.slashRatio = tm.SlashRatio
	return nil
}

func stringToUint256(value string) (*uint256.Int, error) {
	if value == "" {
		return nil, nil
	}
	returnValue, err := uint256.FromHex(value)
	if err != nil {
		return nil, err
	}
	return returnValue, nil
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

	return r.minValidatorStake
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

func (r *GovRule) GasPrice() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.gasPrice)
}

func (r *GovRule) MinTrxGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minTrxGas
}

func (r *GovRule) MaxTrxGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxTrxGas
}

func (r *GovRule) MaxBlockGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxBlockGas
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

// utility methods
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
	// 1 VotingPower == 1 RIGO = 10^18 amount
	return new(uint256.Int).Mul(uint256.NewInt(uint64(power)), amountPerPower)
}

func AmountPerPower() *uint256.Int {
	return amountPerPower.Clone()
}

func FeeToGas(fee, price *uint256.Int) uint64 {
	gas := new(uint256.Int).Div(fee, price)
	return gas.Uint64()
}

func GasToFee(gas uint64, price *uint256.Int) *uint256.Int {
	return new(uint256.Int).Mul(uint256.NewInt(gas), price)
}

func MergeGovRule(oldGovRule, newGovRule *GovRule) {
	if newGovRule.version == 0 {
		newGovRule.version = oldGovRule.version
	}

	if newGovRule.maxValidatorCnt == 0 {
		newGovRule.maxValidatorCnt = oldGovRule.maxValidatorCnt
	}

	if newGovRule.minValidatorStake == nil || newGovRule.minValidatorStake.IsZero() {
		newGovRule.minValidatorStake = oldGovRule.minValidatorStake
	}

	if newGovRule.rewardPerPower == nil || newGovRule.rewardPerPower.IsZero() {
		newGovRule.rewardPerPower = oldGovRule.rewardPerPower
	}

	if newGovRule.lazyRewardBlocks == 0 {
		newGovRule.lazyRewardBlocks = oldGovRule.lazyRewardBlocks
	}

	if newGovRule.lazyApplyingBlocks == 0 {
		newGovRule.lazyApplyingBlocks = oldGovRule.lazyApplyingBlocks
	}

	if newGovRule.gasPrice == nil || newGovRule.gasPrice.IsZero() {
		newGovRule.gasPrice = oldGovRule.gasPrice
	}

	if newGovRule.minTrxGas == 0 {
		newGovRule.minTrxGas = oldGovRule.minTrxGas
	}

	if newGovRule.maxTrxGas == 0 {
		newGovRule.maxTrxGas = oldGovRule.maxTrxGas
	}

	if newGovRule.maxBlockGas == 0 {
		newGovRule.maxBlockGas = oldGovRule.maxBlockGas
	}

	if newGovRule.minVotingPeriodBlocks == 0 {
		newGovRule.minVotingPeriodBlocks = oldGovRule.minVotingPeriodBlocks
	}

	if newGovRule.maxVotingPeriodBlocks == 0 {
		newGovRule.maxVotingPeriodBlocks = oldGovRule.maxVotingPeriodBlocks
	}

	if newGovRule.minSelfStakeRatio == 0 {
		newGovRule.minSelfStakeRatio = oldGovRule.minSelfStakeRatio
	}

	if newGovRule.maxUpdatableStakeRatio == 0 {
		newGovRule.maxUpdatableStakeRatio = oldGovRule.maxUpdatableStakeRatio
	}

	if newGovRule.slashRatio == 0 {
		newGovRule.slashRatio = oldGovRule.slashRatio
	}
}

var _ ledger.ILedgerItem = (*GovRule)(nil)
var _ IGovHandler = (*GovRule)(nil)
