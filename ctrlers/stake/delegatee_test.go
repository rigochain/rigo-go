package stake_test

import (
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

var (
	delegatee *stake.Delegatee
)

func TestAppendStake(t *testing.T) {
	delegatee = stake.NewDelegatee(Wallets[0].Address(), Wallets[0].GetPubKey())

	amt0 := bytes.RandBigIntN(govHelper.MaxStakeAmount())
	power0 := govHelper.AmountToPower(amt0)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                        // amount
			bytes.RandInt63n(1_000_000), // height
			bytes.RandHexBytes(32),      //txhash
			govHelper,
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, amt0.String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, power0, delegatee.GetTotalPower())

	from1 := types.RandAddress()
	amt1 := bytes.RandBigIntN(govHelper.MaxStakeAmount())
	power1 := govHelper.AmountToPower(amt1)
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			from1,
			delegatee.Addr,
			amt1,                        // amount
			bytes.RandInt63n(1_000_000), // height
			bytes.RandHexBytes(32),      //txhash
			govHelper,
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, amt1.String(), delegatee.SumAmountOf(from1).String())
	require.Equal(t, power1, delegatee.SumPowerOf(from1))
	require.Equal(t, new(big.Int).Add(amt0, amt1).String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0+power1, delegatee.GetTotalPower())

}

func TestApplyReward(t *testing.T) {
	delegatee := stake.NewDelegatee(Wallets[1].Address(), Wallets[1].GetPubKey())

	// staking
	amt0 := bytes.RandBigIntN(govHelper.MaxStakeAmount())
	delegatee.AddStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                        // amount
			bytes.RandInt63n(1_000_000), // height
			bytes.RandHexBytes(32),      //txhash
			govHelper,
		),
	)
	// first reward
	reward0 := delegatee.DoReward()
	require.Equal(t, reward0, delegatee.GetRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetRewardAmount())

	// second reward
	reward1 := delegatee.DoReward()
	require.Equal(t, reward0, reward1)
	require.Equal(t, new(big.Int).Add(reward0, reward1), delegatee.GetRewardAmount())
	require.Equal(t, delegatee.SumBlockReward(), delegatee.GetRewardAmount())
	require.True(t, delegatee.GetRewardAmount().Sign() > 0)
}

func BenchmarkApplyReward(b *testing.B) {
	delegatee := stake.NewDelegatee(Wallets[1].Address(), Wallets[1].GetPubKey())
	for i := 0; i < 10000; i++ {
		// staking
		power := bytes.RandInt63n(1000000000)
		delegatee.AddStake(
			stake.NewStakeWithPower(
				types.RandAddress(),
				delegatee.Addr,
				power,                       // power
				bytes.RandInt63n(1_000_000), // height
				bytes.RandHexBytes(32),      //txhash
				govHelper,
			),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := delegatee.DoReward()
		require.True(b, rewarded.Sign() > 0)
	}
}
