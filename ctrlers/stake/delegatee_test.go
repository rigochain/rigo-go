package stake_test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"math/rand"
	"testing"
)

func TestSlash(t *testing.T) {
	for _, w := range Wallets[:21] {
		delegatee := stake.NewDelegatee(w.Address(), w.GetPubKey())

		selfPower0 := int64(0)
		totalPower0 := int64(0)
		stakesCnt := int64(tmrand.Intn(100) + 100)

		for i := int64(0); i < stakesCnt; i++ {
			stakeAmt := new(uint256.Int).Mul(govHelper.AmountPerPower(), uint256.NewInt(uint64(tmrand.Int63n(100_000_000))))
			stakePower := govHelper.AmountToPower(stakeAmt)

			totalPower0 += stakePower
			fromAddr := types.RandAddress()
			if tmrand.Int()%2 == 0 {
				selfPower0 += stakePower
				fromAddr = delegatee.Addr
			}

			delegatee.AddStake(
				stake.NewStakeWithAmount(
					fromAddr,
					delegatee.Addr,
					stakeAmt,               // amount
					i+1,                    // height
					bytes.RandHexBytes(32), //txhash
					govHelper,
				),
			)
		}

		blocks := int64(100)
		totalReward0 := uint256.NewInt(0)
		for i := stakesCnt; i < (stakesCnt + blocks); i++ {
			r := delegatee.DoReward(i + 1)
			require.Greater(t, r.Sign(), 0, fmt.Sprintf("at height %d", i))
			_ = totalReward0.Add(totalReward0, r)
		}

		require.Equal(t, totalReward0, delegatee.TotalRewardAmount)
		require.Equal(t, totalReward0, delegatee.SumBlockRewardOf(nil))
		selfReward0 := delegatee.SumBlockRewardOf(delegatee.Addr)

		require.Equal(t, selfPower0, delegatee.SelfPower)
		require.Equal(t, govHelper.PowerToAmount(selfPower0), delegatee.SumAmountOf(delegatee.Addr))
		require.Equal(t, govHelper.PowerToAmount(selfPower0), delegatee.SelfAmount)
		require.Equal(t, totalPower0, delegatee.TotalPower)
		require.Equal(t, govHelper.PowerToAmount(totalPower0), delegatee.SumAmountOf(nil))
		require.Equal(t, govHelper.PowerToAmount(totalPower0), delegatee.TotalAmount)

		delegatee.DoSlash(govHelper.SlashRatio(), govHelper.AmountPerPower(), govHelper.RewardPerPower())

		// calculate slashed power
		_slashed := uint256.NewInt(uint64(selfPower0))
		_ = _slashed.Mul(_slashed, uint256.NewInt(uint64(govHelper.SlashRatio())))
		_ = _slashed.Div(_slashed, uint256.NewInt(uint64(100)))
		slashPower := int64(_slashed.Uint64())
		slashRewardUnit := govHelper.PowerToReward(slashPower)

		expectedSelfReward := new(uint256.Int).Sub(selfReward0, new(uint256.Int).Mul(slashRewardUnit, uint256.NewInt(uint64(blocks))))
		require.Equal(t, expectedSelfReward, delegatee.SumBlockRewardOf(delegatee.Addr))
		expectedTotalReward := new(uint256.Int).Sub(totalReward0, new(uint256.Int).Sub(selfReward0, expectedSelfReward))
		require.Equal(t, expectedTotalReward, delegatee.TotalRewardAmount)
		require.Equal(t, expectedTotalReward, delegatee.SumBlockRewardOf(nil))

		require.Equal(t, selfPower0-slashPower, delegatee.SelfPower)
		require.Equal(t, govHelper.PowerToAmount(selfPower0-slashPower), delegatee.SelfAmount)
		require.Equal(t, govHelper.PowerToAmount(selfPower0-slashPower), delegatee.SumAmountOf(delegatee.Addr))
		require.Equal(t, totalPower0-slashPower, delegatee.TotalPower)
		require.Equal(t, govHelper.PowerToAmount(totalPower0-slashPower), delegatee.TotalAmount)
		require.Equal(t, govHelper.PowerToAmount(totalPower0-slashPower), delegatee.SumAmountOf(nil))

		for i, s0 := range delegatee.Stakes {
			expectedRewardUnit := govHelper.PowerToReward(s0.Power)
			require.Equal(t, expectedRewardUnit, delegatee.Stakes[i].BlockRewardUnit)
		}

	}
}

func TestAddStake(t *testing.T) {
	delegatee := stake.NewDelegatee(Wallets[0].Address(), Wallets[0].GetPubKey())

	amt0 := bytes.RandU256IntN(govHelper.MaxStakeAmount())
	power0 := govHelper.AmountToPower(amt0)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                   // amount
			rand.Int63n(1_000_000), // height
			bytes.RandHexBytes(32), //txhash
			govHelper,
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, amt0.String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, power0, delegatee.GetTotalPower())

	from1 := types.RandAddress()
	amt1 := bytes.RandU256IntN(govHelper.MaxStakeAmount())
	power1 := govHelper.AmountToPower(amt1)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			from1,
			delegatee.Addr,
			amt1,                   // amount
			rand.Int63n(1_000_000), // height
			bytes.RandHexBytes(32), //txhash
			govHelper,
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, amt1.String(), delegatee.SumAmountOf(from1).String())
	require.Equal(t, power1, delegatee.SumPowerOf(from1))
	require.Equal(t, new(uint256.Int).Add(amt0, amt1).String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0+power1, delegatee.GetTotalPower())

}

func TestDoReward_Delegatee(t *testing.T) {
	delegatee := stake.NewDelegatee(Wallets[1].Address(), Wallets[1].GetPubKey())

	// staking
	height0 := rand.Int63n(1_000_000)
	amt0 := bytes.RandU256IntN(govHelper.MaxStakeAmount())
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                   // amount
			height0+1,              // height
			bytes.RandHexBytes(32), //txhash
			govHelper,
		),
	)
	// not rewarded - height0 is not enough
	reward0 := delegatee.DoReward(height0)
	require.Equal(t, 0, delegatee.GetTotalRewardAmount().Sign())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetTotalRewardAmount())

	// first reward
	reward1 := delegatee.DoReward(height0 + 1)
	require.Equal(t, new(uint256.Int).Add(reward0, reward1), delegatee.GetTotalRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetTotalRewardAmount())
	require.True(t, delegatee.GetTotalRewardAmount().Sign() > 0)

	// first reward
	reward2 := delegatee.DoReward(height0 + 2)
	require.Equal(t, reward1, reward2)
	require.Equal(t, new(uint256.Int).Add(reward1, reward2), delegatee.GetTotalRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetTotalRewardAmount())
	require.True(t, delegatee.GetTotalRewardAmount().Sign() > 0)
}

func BenchmarkApplyReward(b *testing.B) {
	delegatee := stake.NewDelegatee(Wallets[1].Address(), Wallets[1].GetPubKey())
	for i := 0; i < 10000; i++ {
		// staking
		power := rand.Int63n(1000000000)
		delegatee.AddStake(
			stake.NewStakeWithPower(
				types.RandAddress(),
				delegatee.Addr,
				power,                  // power
				rand.Int63n(1),         // height
				bytes.RandHexBytes(32), //txhash
				govHelper,
			),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := delegatee.DoReward(int64(i + 1))
		require.True(b, rewarded.Sign() > 0)
	}
}
