package test

import (
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestNonce(t *testing.T) {
	require.NoError(t, W0.SyncAccount(rweb3))
	require.NoError(t, W0.Unlock(TESTPASS))
	ret, err := W0.TransferSync(W1.Address(), gas, amt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)
	txHash := ret.Hash

	_, err = waitTrxResult(txHash, 10)
	require.NoError(t, err)

	// do not add/sync nonce

	ret, err = W0.TransferSync(W1.Address(), gas, amt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrInvalidNonce.Error(), ret.Log)

	// return fund
	require.NoError(t, W1.SyncAccount(rweb3))

	bal := W1.GetBalance()
	amt0 := new(big.Int).Sub(bal, gas)

	ret, err = W1.TransferSync(W0.Address(), gas, amt0, rweb3)
	require.NoError(t, err) // not insufficient fund
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)
	txHash = ret.Hash

	_, err = waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.NoError(t, W1.SyncAccount(rweb3))
	require.Equal(t, big.NewInt(0), W1.GetBalance())
}
