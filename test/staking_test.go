package test

import (
	"bytes"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestStakingToValidator(t *testing.T) {

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt := abytes.RandBigIntN(bal)

	require.NoError(t, w.Unlock(TESTPASS))
	// self staking
	ret, err := w.StakingSync(validatorWallet.Address(), gas, stakeAmt, rweb3)
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

func TestStakingToSelf(t *testing.T) {

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt := abytes.RandBigIntN(bal)

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
}
