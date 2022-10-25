package stake_test

import (
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

var (
	delegatee *stake.Delegatee
)

func TestAppendStake(t *testing.T) {
	delegatee = stake.NewDelegatee(Wallets[0].GetAddress(), Wallets[0].W.GetPubKey())

	amt0 := libs.RandBigIntN(testGovRules.MaxStakeAmount())
	power0 := testGovRules.AmountToPower(amt0)
	delegatee.AppendStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                       // amount
			libs.RandInt63n(1_000_000), // height
			libs.RandHexBytes(32),      //txhash
			testGovRules,
		),
	)

	require.Equal(t, amt0.String(), delegatee.GetSelfAmount().String())
	require.Equal(t, power0, delegatee.GetSelfPower())
	require.Equal(t, amt0.String(), delegatee.GetTotalAmount().String())
	require.Equal(t, power0, delegatee.GetTotalPower())

	from1 := libs.RandAddress()
	amt1 := libs.RandBigIntN(testGovRules.MaxStakeAmount())
	power1 := testGovRules.AmountToPower(amt1)
	delegatee.AppendStake(
		stake.NewStakeWithAmount(
			from1,
			delegatee.Addr,
			amt1,                       // amount
			libs.RandInt63n(1_000_000), // height
			libs.RandHexBytes(32),      //txhash
			testGovRules,
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
	delegatee := stake.NewDelegatee(Wallets[1].GetAddress(), Wallets[1].W.GetPubKey())

	// staking
	amt0 := libs.RandBigIntN(testGovRules.MaxStakeAmount())
	delegatee.AppendStake(
		stake.NewStakeWithAmount(
			delegatee.Addr,
			delegatee.Addr,
			amt0,                       // amount
			libs.RandInt63n(1_000_000), // height
			libs.RandHexBytes(32),      //txhash
			testGovRules,
		),
	)
	// first reward
	reward0 := delegatee.ApplyReward()
	require.Equal(t, reward0, delegatee.GetTotalReward())
	require.Equal(t, delegatee.SumReward(), delegatee.GetTotalReward())

	// second reward
	reward1 := delegatee.ApplyReward()
	require.Equal(t, reward0, reward1)
	require.Equal(t, new(big.Int).Add(reward0, reward1), delegatee.GetTotalReward())
	require.Equal(t, delegatee.SumReward(), delegatee.GetTotalReward())
	require.True(t, delegatee.GetTotalReward().Sign() > 0)
}

func BenchmarkApplyReward(b *testing.B) {
	delegatee := stake.NewDelegatee(Wallets[1].GetAddress(), Wallets[1].W.GetPubKey())
	for i := 0; i < 10000; i++ {
		// staking
		power := libs.RandInt63n(1000000000)
		delegatee.AppendStake(
			stake.NewStakeWithPower(
				libs.RandAddress(),
				delegatee.Addr,
				power,                      // power
				libs.RandInt63n(1_000_000), // height
				libs.RandHexBytes(32),      //txhash
				testGovRules,
			),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rewarded := delegatee.ApplyReward()
		require.True(b, rewarded.Sign() > 0)
	}
}
