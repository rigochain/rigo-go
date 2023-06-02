package stake_test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"math/rand"
	"testing"
)

func TestSlashAll(t *testing.T) {
	for _, w := range Wallets[:21] {
		delegatee := stake.NewDelegatee(w.Address(), w.GetPubKey())

		selfPower0 := int64(0)
		totalPower0 := int64(0)
		stakesCnt := int64(tmrand.Intn(100) + 100)

		for i := int64(0); i < stakesCnt; i++ {
			stakeAmt := new(uint256.Int).Mul(ctrlertypes.AmountPerPower(), uint256.NewInt(uint64(tmrand.Int63n(100_000_000))))
			stakePower := ctrlertypes.AmountToPower(stakeAmt)

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
				),
			)
		}

		blocks := int64(100)
		totalReward0 := uint256.NewInt(0)
		for i := stakesCnt; i < (stakesCnt + blocks); i++ {
			r := delegatee.DoReward(i+1, ctrlertypes.AmountPerPower(), govParams.RewardPerPower())
			require.Greater(t, r.Sign(), 0, fmt.Sprintf("at height %d", i))
			_ = totalReward0.Add(totalReward0, r)
		}
		require.Equal(t, totalReward0, delegatee.TotalRewardAmount)
		require.Equal(t, totalReward0, delegatee.SumBlockRewardOf(nil))

		require.Equal(t, selfPower0, delegatee.SelfPower)
		require.Equal(t, ctrlertypes.PowerToAmount(selfPower0), delegatee.SumAmountOf(delegatee.Addr))
		require.Equal(t, ctrlertypes.PowerToAmount(selfPower0), delegatee.SelfAmount)
		require.Equal(t, totalPower0, delegatee.TotalPower)
		require.Equal(t, ctrlertypes.PowerToAmount(totalPower0), delegatee.SumAmountOf(nil))
		require.Equal(t, ctrlertypes.PowerToAmount(totalPower0), delegatee.TotalAmount)

		oriStakes := make([]*stake.Stake, delegatee.StakesLen())
		for i, s0 := range delegatee.GetAllStakes() {
			oriStakes[i] = s0.Clone()
		}

		totalSlashedPower0 := delegatee.DoSlash(govParams.SlashRatio(), ctrlertypes.AmountPerPower(), govParams.RewardPerPower(), true)

		expectedTotalSlashedPower0 := int64(0)
		expectedTotalReceivedReward := uint256.NewInt(0)
		for _, s0 := range oriStakes {
			p0 := uint256.NewInt(uint64(s0.Power))
			_ = p0.Mul(p0, uint256.NewInt(uint64(govParams.SlashRatio())))
			_ = p0.Div(p0, uint256.NewInt(uint64(100)))
			slashedPower := int64(p0.Uint64())
			if slashedPower < 1 {
				slashedPower = s0.Power
				expectedTotalSlashedPower0 += slashedPower
				continue
			}

			expectedPower := s0.Power - slashedPower
			require.NotEqual(t, expectedPower, s0.Power)

			expectedAmt := new(uint256.Int).Mul(ctrlertypes.AmountPerPower(), uint256.NewInt(uint64(expectedPower)))
			require.NotEqual(t, expectedAmt.Dec(), s0.Amount.Dec())

			expectedReceivedReward := new(uint256.Int).Mul(s0.ReceivedReward, uint256.NewInt(uint64(govParams.SlashRatio())))
			_ = expectedReceivedReward.Div(expectedReceivedReward, uint256.NewInt(uint64(100)))
			require.NotEqual(t, expectedReceivedReward.Dec(), s0.ReceivedReward.Dec())

			_, s1 := delegatee.FindStake(s0.TxHash)

			require.NotNil(t, s1)
			require.Equal(t, s0.TxHash, s1.TxHash)
			require.Equal(t, expectedPower, s1.Power)
			require.Equal(t, expectedAmt.Dec(), s1.Amount.Dec())
			require.Equal(t, expectedReceivedReward.Dec(), s1.ReceivedReward.Dec())

			expectedTotalSlashedPower0 += slashedPower
			_ = expectedTotalReceivedReward.Add(expectedTotalReceivedReward, expectedReceivedReward)
		}
		require.Equal(t, expectedTotalSlashedPower0, totalSlashedPower0)
		require.Equal(t, totalPower0-totalSlashedPower0, delegatee.TotalPower)
		require.Equal(t, totalPower0-totalSlashedPower0, delegatee.SumPowerOf(nil))
		require.Equal(t, expectedTotalReceivedReward.Dec(), delegatee.TotalRewardAmount.Dec())
	}
}

func TestAddStake(t *testing.T) {
	delegatee := stake.NewDelegatee(Wallets[0].Address(), Wallets[0].GetPubKey())

	amt0 := bytes.RandU256IntN(ctrlertypes.MaxStakeAmount())
	power0 := ctrlertypes.AmountToPower(amt0)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                   // amount
			rand.Int63n(1_000_000), // height
			bytes.RandHexBytes(32), //txhash
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, amt0.String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, power0, delegatee.GetTotalPower())

	from1 := types.RandAddress()
	amt1 := bytes.RandU256IntN(ctrlertypes.MaxStakeAmount())
	power1 := ctrlertypes.AmountToPower(amt1)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			from1,
			delegatee.Addr,
			amt1,                   // amount
			rand.Int63n(1_000_000), // height
			bytes.RandHexBytes(32), //txhash
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
	amt0 := bytes.RandU256IntN(ctrlertypes.MaxStakeAmount())
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                   // amount
			height0+1,              // height
			bytes.RandHexBytes(32), //txhash
		),
	)
	// not rewarded - height0 is not enough
	reward0 := delegatee.DoReward(height0, ctrlertypes.AmountPerPower(), govParams.RewardPerPower())
	require.Equal(t, 0, delegatee.GetTotalRewardAmount().Sign())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetTotalRewardAmount())

	// first reward
	reward1 := delegatee.DoReward(height0+1, ctrlertypes.AmountPerPower(), govParams.RewardPerPower())
	require.Equal(t, new(uint256.Int).Add(reward0, reward1), delegatee.GetTotalRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetTotalRewardAmount())
	require.True(t, delegatee.GetTotalRewardAmount().Sign() > 0)

	// first reward
	reward2 := delegatee.DoReward(height0+2, ctrlertypes.AmountPerPower(), govParams.RewardPerPower())
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
			),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := delegatee.DoReward(int64(i+1), ctrlertypes.AmountPerPower(), govParams.RewardPerPower())
		require.True(b, rewarded.Sign() > 0)
	}
}
