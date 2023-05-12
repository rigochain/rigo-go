package stake_test

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

var (
	delegatee *stake.Delegatee
)

func TestAppendStake(t *testing.T) {
	delegatee = stake.NewDelegatee(Wallets[0].Address(), Wallets[0].GetPubKey())

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

func TestApplyReward(t *testing.T) {
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
	require.Equal(t, 0, delegatee.GetRewardAmount().Sign())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetRewardAmount())

	// first reward
	reward1 := delegatee.DoReward(height0 + 1)
	require.Equal(t, new(uint256.Int).Add(reward0, reward1), delegatee.GetRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetRewardAmount())
	require.True(t, delegatee.GetRewardAmount().Sign() > 0)

	// first reward
	reward2 := delegatee.DoReward(height0 + 2)
	require.Equal(t, reward1, reward2)
	require.Equal(t, new(uint256.Int).Add(reward1, reward2), delegatee.GetRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetRewardAmount())
	require.True(t, delegatee.GetRewardAmount().Sign() > 0)
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
