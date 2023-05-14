package test

import (
	"bytes"
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3"
	rtypes0 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

//
// DO NOT RUN this test code yet.

func TestQueryValidators(t *testing.T) {
	rweb3 := randRigoWeb3()

	ret, err := queryValidators(1, rweb3)

	require.NoError(t, err)
	require.Equal(t, len(validatorWallets), len(ret.Validators))
	for _, val := range ret.Validators {
		require.True(t, isValidator(rtypes0.Address(val.Address)))
	}
}

func TestMinSelfStakeRatio(t *testing.T) {
	rweb3 := randRigoWeb3()

	valWal := validatorWallets[0]
	valStakes, err := rweb3.GetDelegatee(valWal.Address())
	require.NoError(t, err)

	sender := randCommonWallet()
	require.NoError(t, sender.Unlock(defaultRpcNode.Pass))
	require.NoError(t, sender.SyncAccount(rweb3))

	// allowed
	maxAllowedAmt := valStakes.TotalAmount
	ret, err := sender.StakingSync(valWal.Address(), gas, maxAllowedAmt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)

	sender.AddNonce()

	// disallowed
	ret, err = sender.StakingSync(valWal.Address(), gas, uint256.NewInt(1000000000000000000), rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	require.True(t, strings.Contains(ret.Log, "not enough self power"), ret.Log)

}

func TestInvalidStakeAmount(t *testing.T) {
	rweb3 := randRigoWeb3()

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))
	require.NoError(t, w.Unlock(defaultRpcNode.Pass))

	// too small
	stakeAmt := uint256.MustFromDecimal("1111")

	ret, err := w.StakingSync(w.Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	// not multiple
	stakeAmt = uint256.MustFromDecimal("1000000000000000001")

	ret, err = w.StakingSync(w.Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
}

func TestDelegating(t *testing.T) {
	rweb3 := randRigoWeb3()

	var w *web3.Wallet
	for {
		w = randCommonWallet()
		require.NoError(t, w.SyncAccount(rweb3))
		if w.GetBalance().Sign() > 0 {
			break
		}
	}

	vals, err := queryValidators(0, rweb3)
	require.NoError(t, err)
	fmt.Println("query validator power", vals.Validators[0].VotingPower)
	valAddr := rtypes0.Address(vals.Validators[0].Address)

	valStakes0, err := rweb3.GetDelegatee(valAddr)
	require.NoError(t, err)

	stakeAmt := uint256.NewInt(1000000000000000000) // 10^18
	stakePower := int64(1)

	require.NoError(t, w.Unlock(defaultRpcNode.Pass))
	// self staking
	ret, err := w.StakingSync(valAddr, gas, stakeAmt, rweb3)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 30, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	require.Equal(t, gas, txRet.TxDetail.Gas)
	require.Equal(t, stakeAmt, txRet.TxDetail.Amount)

	// check stakes
	found := false
	stakes, err := rweb3.GetStakes(w.Address())
	require.NoError(t, err)
	require.True(t, len(stakes) > 0)
	for _, s0 := range stakes {
		if bytes.Compare(s0.TxHash, txHash) == 0 {
			require.Equal(t, stakeAmt, s0.Amount)
			found = true
		}
	}
	require.True(t, found)

	valStakes1, err := rweb3.GetDelegatee(valAddr)
	require.NoError(t, err)

	require.Equal(t, valStakes0.SelfPower, valStakes1.SelfPower)
	require.Equal(t, valStakes0.TotalPower+stakePower, valStakes1.TotalPower)
	require.Equal(t, new(uint256.Int).Add(valStakes0.TotalAmount, stakeAmt), valStakes1.TotalAmount)

	time.Sleep(5 * time.Second)

	vals, err = queryValidators(0, rweb3)
	require.NoError(t, err)
	fmt.Println("query validator power", vals.Validators[0].VotingPower)
}
