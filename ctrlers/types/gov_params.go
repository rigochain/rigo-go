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

type GovParams struct {
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
	signedBlocksWindow     int64
	minSignedRatio         int64

	mtx sync.RWMutex
}

func DefaultGovParams() *GovParams {
	return &GovParams{
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
		minTrxGas:              uint64(100_000),                // 0.001 RIGO = 10^15 wei
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  259200,  // = 60 * 60 * 24 * 3 => 3 days
		maxVotingPeriodBlocks:  2592000, // = 60 * 60 * 24 * 30 => 30 days
		minSelfStakeRatio:      50,      // 50%
		maxUpdatableStakeRatio: 30,      // 30%
		slashRatio:             50,      // 50%
		signedBlocksWindow:     10000,   // 10000 blocks
		minSignedRatio:         5,       // 5%
	}
}

func Test1GovParams() *GovParams {
	return &GovParams{
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
		minSelfStakeRatio:      50,    // 50%
		maxUpdatableStakeRatio: 30,    // 30%
		slashRatio:             50,    // 50%
		signedBlocksWindow:     10000, // 10000 blocks
		minSignedRatio:         5,     // 5%
	}
}

func Test2GovParams() *GovParams {
	return &GovParams{
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
		minSelfStakeRatio:      50,    // 50%
		maxUpdatableStakeRatio: 30,    // 30%
		slashRatio:             50,    // 50%
		signedBlocksWindow:     10000, // 10000 blocks
		minSignedRatio:         5,     // 5%
	}
}

func Test3GovParams() *GovParams {
	return &GovParams{
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
		signedBlocksWindow:     10000,
		minSignedRatio:         5,
	}
}

func Test4GovParams() *GovParams {
	return &GovParams{
		version:                4,
		maxValidatorCnt:        13,
		minValidatorStake:      uint256.MustFromDecimal("7000000000000000000000000"),
		rewardPerPower:         uint256.NewInt(4_756_468_797),
		lazyRewardBlocks:       20,
		lazyApplyingBlocks:     259200,
		gasPrice:               uint256.NewInt(10_000_000_000),
		minTrxGas:              uint64(100_000),
		maxTrxGas:              math.MaxUint64,
		maxBlockGas:            math.MaxUint64,
		minVotingPeriodBlocks:  259200,
		maxVotingPeriodBlocks:  2592000,
		minSelfStakeRatio:      50,
		maxUpdatableStakeRatio: 10,
		slashRatio:             50,
		signedBlocksWindow:     10000,
		minSignedRatio:         5,
	}
}

func Test5GovParams() *GovParams {
	return &GovParams{
		version:                3,
		minValidatorStake:      uint256.MustFromDecimal("0"),
		minSelfStakeRatio:      40,
		maxUpdatableStakeRatio: 50,
		slashRatio:             60,
	}
}

func DecodeGovParams(bz []byte) (*GovParams, xerrors.XError) {
	ret := &GovParams{}
	if xerr := ret.Decode(bz); xerr != nil {
		return nil, xerr
	}
	return ret, nil
}

func (r *GovParams) Key() ledger.LedgerKey {
	return ledger.ToLedgerKey(bytes.ZeroBytes(32))
}

func (r *GovParams) Decode(bz []byte) xerrors.XError {
	pm := &GovParamsProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.From(err)
	}
	r.fromProto(pm)
	return nil
}

func (r *GovParams) Encode() ([]byte, xerrors.XError) {
	if bz, err := proto.Marshal(r.toProto()); err != nil {
		return nil, xerrors.From(err)
	} else {
		return bz, nil
	}
}

func (r *GovParams) fromProto(pm *GovParamsProto) {
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
	r.signedBlocksWindow = pm.SignedBlocksWindow
	r.minSignedRatio = pm.MinSignedRatio
}

func (r *GovParams) toProto() *GovParamsProto {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	a := &GovParamsProto{
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
		SignedBlocksWindow:     r.signedBlocksWindow,
		MinSignedRatio:         r.minSignedRatio,
	}
	return a
}

func (r *GovParams) MarshalJSON() ([]byte, error) {
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
		MaxTrxGas              uint64 `json:"maxTrxGas"`
		MaxBlockGas            uint64 `json:"maxBlockGas"`
		MinVotingBlocks        int64  `json:"minVotingPeriodBlocks"`
		MaxVotingBlocks        int64  `json:"maxVotingPeriodBlocks"`
		MinSelfStakeRatio      int64  `json:"minSelfStakeRatio"`
		MaxUpdatableStakeRatio int64  `json:"maxUpdatableStakeRatio"`
		SlashRatio             int64  `json:"slashRatio"`
		SignedBlocksWindow     int64  `json:"signedBlocksWindow"`
		MinSignedRatio         int64  `json:"minSignedRatio"`
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
		SignedBlocksWindow:     r.signedBlocksWindow,
		MinSignedRatio:         r.minSignedRatio,
	}
	return tmjson.Marshal(tm)
}

func uint256ToString(value *uint256.Int) string {
	if value == nil {
		return ""
	}
	return value.Dec()
}

func (r *GovParams) UnmarshalJSON(bz []byte) error {
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
		SignedBlocksWindow     int64  `json:"signedBlocksWindow"`
		MinSignedRatio         int64  `json:"minSignedRatio"`
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
	r.signedBlocksWindow = tm.SignedBlocksWindow
	r.minSignedRatio = tm.MinSignedRatio
	return nil
}

func stringToUint256(value string) (*uint256.Int, error) {
	if value == "" {
		return nil, nil
	}
	returnValue, err := uint256.FromDecimal(value)
	if err != nil {
		return nil, err
	}
	return returnValue, nil
}

func (r *GovParams) Version() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.version
}

func (r *GovParams) MaxValidatorCnt() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxValidatorCnt
}

func (r *GovParams) MinValidatorStake() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minValidatorStake
}

func (r *GovParams) RewardPerPower() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.rewardPerPower)
}

func (r *GovParams) LazyRewardBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.lazyRewardBlocks
}

func (r *GovParams) LazyApplyingBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.lazyApplyingBlocks
}

func (r *GovParams) GasPrice() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Set(r.gasPrice)
}

func (r *GovParams) MinTrxGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minTrxGas
}

func (r *GovParams) MinTrxFee() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Mul(uint256.NewInt(r.minTrxGas), r.gasPrice)
}

func (r *GovParams) MaxTrxGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxTrxGas
}

func (r *GovParams) MaxTrxFee() *uint256.Int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return new(uint256.Int).Mul(uint256.NewInt(r.maxTrxGas), r.gasPrice)
}

func (r *GovParams) MaxBlockGas() uint64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxBlockGas
}

func (r *GovParams) MinVotingPeriodBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minVotingPeriodBlocks
}

func (r *GovParams) MaxVotingPeriodBlocks() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxVotingPeriodBlocks
}
func (r *GovParams) MinSelfStakeRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minSelfStakeRatio
}
func (r *GovParams) MaxUpdatableStakeRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.maxUpdatableStakeRatio
}
func (r *GovParams) SlashRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.slashRatio
}

func (r *GovParams) SignedBlocksWindow() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.signedBlocksWindow
}

func (r *GovParams) MinSignedRatio() int64 {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return r.minSignedRatio
}

func (r *GovParams) String() string {
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

func MergeGovParams(oldParams, newParams *GovParams) {
	if newParams.version == 0 {
		newParams.version = oldParams.version
	}

	if newParams.maxValidatorCnt == 0 {
		newParams.maxValidatorCnt = oldParams.maxValidatorCnt
	}

	if newParams.minValidatorStake == nil || newParams.minValidatorStake.IsZero() {
		newParams.minValidatorStake = oldParams.minValidatorStake
	}

	if newParams.rewardPerPower == nil || newParams.rewardPerPower.IsZero() {
		newParams.rewardPerPower = oldParams.rewardPerPower
	}

	if newParams.lazyRewardBlocks == 0 {
		newParams.lazyRewardBlocks = oldParams.lazyRewardBlocks
	}

	if newParams.lazyApplyingBlocks == 0 {
		newParams.lazyApplyingBlocks = oldParams.lazyApplyingBlocks
	}

	if newParams.gasPrice == nil || newParams.gasPrice.IsZero() {
		newParams.gasPrice = oldParams.gasPrice
	}

	if newParams.minTrxGas == 0 {
		newParams.minTrxGas = oldParams.minTrxGas
	}

	if newParams.maxTrxGas == 0 {
		newParams.maxTrxGas = oldParams.maxTrxGas
	}

	if newParams.maxBlockGas == 0 {
		newParams.maxBlockGas = oldParams.maxBlockGas
	}

	if newParams.minVotingPeriodBlocks == 0 {
		newParams.minVotingPeriodBlocks = oldParams.minVotingPeriodBlocks
	}

	if newParams.maxVotingPeriodBlocks == 0 {
		newParams.maxVotingPeriodBlocks = oldParams.maxVotingPeriodBlocks
	}

	if newParams.minSelfStakeRatio == 0 {
		newParams.minSelfStakeRatio = oldParams.minSelfStakeRatio
	}

	if newParams.maxUpdatableStakeRatio == 0 {
		newParams.maxUpdatableStakeRatio = oldParams.maxUpdatableStakeRatio
	}

	if newParams.slashRatio == 0 {
		newParams.slashRatio = oldParams.slashRatio
	}

	if newParams.signedBlocksWindow == 0 {
		newParams.signedBlocksWindow = oldParams.signedBlocksWindow
	}

	if newParams.minSignedRatio == 0 {
		newParams.minSignedRatio = oldParams.minSignedRatio
	}
}

var _ ledger.ILedgerItem = (*GovParams)(nil)
var _ IGovHandler = (*GovParams)(nil)
