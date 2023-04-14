package test

import (
	"bytes"
	rtypes0 "github.com/rigochain/rigo-go/types"
	rbytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestQueryValidators(t *testing.T) {
	ret, err := rweb3.GetValidators(1, 1, len(validatorWallets))

	require.NoError(t, err)
	require.Equal(t, len(validatorWallets), len(ret.Validators))
	for _, val := range ret.Validators {
		require.True(t, isValidator(rtypes0.Address(val.Address)))
	}
}

func TestStakingToSelf(t *testing.T) {

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt, ok := new(big.Int).SetString("4000000000000000000", 10) //rbytes.RandBigIntN(bal)
	require.True(t, ok)

	require.NoError(t, w.Unlock(TESTPASS))
	// self staking
	ret, err := w.StakingSync(w.Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	require.Equal(t, gas, txRet.TxDetail.Gas)
	require.Equal(t, stakeAmt, txRet.TxDetail.Amount)

	// check stakes
	found := false
	stakes, err := rweb3.GetStakes(w.Address())
	require.True(t, len(stakes) > 0)
	for _, s0 := range stakes {
		if bytes.Compare(s0.TxHash, txHash) == 0 {
			require.Equal(t, stakeAmt, s0.Amount)
			found = true
		}
	}
	require.True(t, found)
	require.NoError(t, err)

	addValidatorWallet(w)

	TestQueryValidators(t)
}

func TestDelegating(t *testing.T) {

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt := rbytes.RandBigIntN(bal)

	require.NoError(t, w.Unlock(TESTPASS))
	// self staking
	ret, err := w.StakingSync(validatorWallets[0].Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	require.Equal(t, gas, txRet.TxDetail.Gas)
	require.Equal(t, stakeAmt, txRet.TxDetail.Amount)

	// check stakes
	found := false
	stakes, err := rweb3.GetStakes(w.Address())
	require.True(t, len(stakes) > 0)
	for _, s0 := range stakes {
		if bytes.Compare(s0.TxHash, txHash) == 0 {
			require.Equal(t, stakeAmt, s0.Amount)
			found = true
		}
	}
	require.True(t, found)
	require.NoError(t, err)
}
