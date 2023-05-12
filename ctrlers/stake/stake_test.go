package stake_test

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

type stakeTestObj struct {
	s              *stake.Stake
	expectedReward *uint256.Int
}

func TestNewStake(t *testing.T) {
	amt := bytes.RandU256IntN(govHelper.MaxStakeAmount())
	s0 := stake.NewStakeWithAmount(
		types.RandAddress(),
		types.RandAddress(),
		amt, 1, nil,
		govHelper)

	require.True(t, s0.Power > int64(0))
	require.Equal(t, govHelper.AmountToPower(amt), s0.Power)
	require.True(t, s0.BlockRewardUnit.Sign() > 0)
	require.Equal(t, govHelper.PowerToReward(s0.Power), s0.BlockRewardUnit)
	require.Equal(t, uint256.NewInt(0), s0.ReceivedReward)
}

func TestApplyRewardByStake(t *testing.T) {
	stakeTestObjs := make([]*stakeTestObj, 1000)

	for i := 0; i < 1000; i++ {
		amt := bytes.RandU256IntN(new(uint256.Int).Div(govHelper.MaxStakeAmount(), uint256.NewInt(1000)))
		txhash := bytes.RandBytes(32)
		stakeTestObjs[i] = &stakeTestObj{
			s: stake.NewStakeWithAmount(
				types.RandAddress(),
				types.RandAddress(),
				amt, 1, txhash,
				govHelper),
			expectedReward: new(uint256.Int),
		}
	}

	// reward for 20000 blocks
	rand.Seed(time.Now().UnixNano())
	blocks := 20000
	for n := 0; n < blocks; n++ {
		i := rand.Int() % len(stakeTestObjs)
		stake := stakeTestObjs[i].s
		stake.ApplyReward(int64(n + 1))
		stakeTestObjs[i].expectedReward.Add(stakeTestObjs[i].expectedReward, stake.BlockRewardUnit)
	}

	for i := 0; i < len(stakeTestObjs); i++ {
		require.Equal(t, stakeTestObjs[i].expectedReward, stakeTestObjs[i].s.ReceivedReward)
	}
}
