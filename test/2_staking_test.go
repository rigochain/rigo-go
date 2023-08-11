package test

import (
	"bytes"
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
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
	for _, val := range ret.Validators {
		require.True(t, isValidator(rtypes0.Address(val.Address)))
	}
}

func TestMinSelfStakeRatio(t *testing.T) {
	rweb3 := randRigoWeb3()

	govRule, err := rweb3.GetRule()
	require.NoError(t, err)

	valWal := validatorWallets[0]
	valStakes, err := rweb3.GetDelegatee(valWal.Address())
	require.NoError(t, err)

	sender := randCommonWallet()
	require.NoError(t, sender.Unlock(defaultRpcNode.Pass))
	require.NoError(t, sender.SyncAccount(rweb3))

	// get allowed delegating
	maxAllowedAmt := ctrlertypes.PowerToAmount(valStakes.TotalPower)
	ret, err := sender.StakingSync(valWal.Address(), baseFee, maxAllowedAmt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	sender.AddNonce()

	// not allowed delegating, because `maxAllowedAmt` is already delegated.
	ret, err = sender.StakingSync(valWal.Address(), baseFee, uint256.NewInt(1_000_000_000_000_000_000), rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	require.True(t, strings.Contains(ret.Log, "not enough self power"), ret.Log)

	// self-staking must be allowed.
	// already stake + new stake >= govRule.MinValidatorStake
	allowedMinStake := new(uint256.Int).Sub(govRule.MinValidatorStake(), ctrlertypes.PowerToAmount(valStakes.SelfPower))
	if allowedMinStake.Sign() <= 0 {
		allowedMinStake = uint256.NewInt(10_000_000_000_000_000_000)
	}

	require.NoError(t, valWal.SyncAccount(rweb3))
	require.NoError(t, valWal.Unlock(defaultRpcNode.Pass))
	ret, err = valWal.StakingSync(valWal.Address(), baseFee, allowedMinStake, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log, allowedMinStake.Dec())

	txRet, err := waitTrxResult(ret.Hash, 30, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log)
}

func TestInvalidStakeAmount(t *testing.T) {
	rweb3 := randRigoWeb3()

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))
	require.NoError(t, w.Unlock(defaultRpcNode.Pass))

	// too small
	stakeAmt := uint256.MustFromDecimal("1111")

	ret, err := w.StakingSync(w.Address(), baseFee, stakeAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	// not multiple
	stakeAmt = uint256.MustFromDecimal("1000000000000000001")

	ret, err = w.StakingSync(w.Address(), baseFee, stakeAmt, rweb3)
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
	//fmt.Println("query validator power", vals.Validators[0].VotingPower)
	valAddr := rtypes0.Address(vals.Validators[0].Address)

	valStakes0, err := rweb3.GetDelegatee(valAddr)
	require.NoError(t, err)

	stakeAmt := uint256.NewInt(1_000_000_000_000_000_000) // 10^18
	stakePower := int64(1)

	require.NoError(t, w.Unlock(defaultRpcNode.Pass))

	// self staking
	ret, err := w.StakingSync(valAddr, baseFee, stakeAmt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 30, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	require.Equal(t, baseFee, txRet.TxDetail.Gas)
	require.Equal(t, stakeAmt, txRet.TxDetail.Amount)

	// check stakes
	found := false
	stakes, err := rweb3.GetStakes(w.Address())
	require.NoError(t, err)
	require.True(t, len(stakes) > 0)
	for _, s0 := range stakes {
		if bytes.Compare(s0.TxHash, txHash) == 0 {
			require.Equal(t, stakeAmt, ctrlertypes.PowerToAmount(s0.Power))
			found = true
		}
	}
	require.True(t, found)

	valStakes1, err := rweb3.GetDelegatee(valAddr)
	require.NoError(t, err)

	require.Equal(t, valStakes0.SelfPower, valStakes1.SelfPower)
	require.Equal(t, valStakes0.TotalPower+stakePower, valStakes1.TotalPower)
	require.Equal(t, new(uint256.Int).Add(ctrlertypes.PowerToAmount(valStakes0.TotalPower), stakeAmt), ctrlertypes.PowerToAmount(valStakes1.TotalPower))

	fmt.Println("Wait 5 seconds...")
	time.Sleep(5 * time.Second)

	vals, err = queryValidators(0, rweb3)
	require.NoError(t, err)
	fmt.Println("query validator power", vals.Validators[0].VotingPower)
}

func TestMinValidatorStake(t *testing.T) {
	rweb3 := randRigoWeb3()

	govRule, err := rweb3.GetRule()
	require.NoError(t, err)

	sender := randCommonWallet()
	require.NoError(t, sender.Unlock(defaultRpcNode.Pass))
	require.NoError(t, sender.SyncAccount(rweb3))

	minValidatorStake := govRule.MinValidatorStake()
	_amt := new(uint256.Int).Sub(minValidatorStake, ctrlertypes.AmountPerPower())
	ret, err := sender.StakingSync(sender.Address(), baseFee, _amt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	require.Contains(t, ret.Log, "too small stake to become validator", ret.Log)
}
