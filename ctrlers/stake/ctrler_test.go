package stake_test

import (
	"fmt"
	"github.com/kysee/arcanus/ctrlers/stake"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

var (
	addr0        = libs.RandAddress()
	dbDir        = filepath.Join(os.TempDir(), "stake-ctrler-test")
	maxAmount, _ = new(big.Int).SetString(types.MXCO, 10)
)

func TestStakerOrder(t *testing.T) {
	os.RemoveAll(dbDir)

	ctrler, err := stake.NewStakeCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)

	var vals stake.StakeSetArray
	for i := 0; i < 50; i++ {
		w0 := client.NewWallet([]byte("1"))
		val0 := ctrler.AddStakerWith(w0.Address(), w0.GetPubKey())
		require.NotNil(t, val0)
		val0.AppendStake(stake.NewStakeWithAmount(w0.Address(), libs.RandBigIntN(maxAmount), int64(i+1), libs.RandBytes(32), govRules))

		vals = append(vals, val0)
	}
	require.Equal(t, 50, ctrler.StakersLen())

	_ = ctrler.UpdateValidators() // sort

	var preVal *stake.StakeSet
	for i := 0; i < ctrler.StakersLen(); i++ {
		staker := ctrler.GetStaker(i)
		if preVal != nil {
			require.Truef(t, preVal.TotalPower >= staker.TotalPower,
				fmt.Sprintf("invalid power(stake) order: pre: %v, curr: %v", preVal.TotalPower, staker.TotalPower))
		}
		preVal = staker
	}
}

func TestStakerReward(t *testing.T) {
	// todo: check if reward is wrong or not
}
