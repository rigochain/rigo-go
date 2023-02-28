package test

import (
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestSyncWallet(t *testing.T) {
	for _, w := range wallets {
		require.NoError(t, w.SyncAccount())
	}
}

func TestTransfer(t *testing.T) {
	require.NoError(t, W0.SyncAccount())
	require.NoError(t, W1.SyncAccount())

	ret, err := W0.TransferSync(W1.Address(), gas, amt)
	require.Error(t, err) // unlock

	require.NoError(t, W0.Unlock(TESTPASS))

	ret, err = W0.TransferSync(W1.Address(), gas, amt)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 10)
	require.NoError(t, err)

	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	//require.Equal(t, trxs.TRX_TRANSFER, txRet.TxDetail.Payload.Type())
	require.Equal(t, gas, txRet.TxDetail.Gas)
	require.Equal(t, amt, txRet.TxDetail.Amount)

	expectedBalance := new(big.Int).Sub(W0.GetBalance(), new(big.Int).Add(amt, gas))
	require.NoError(t, W0.SyncAccount())
	require.Equal(t, expectedBalance, W0.GetBalance())

	expectedBalance = new(big.Int).Add(W1.GetBalance(), amt)
	require.NoError(t, W1.SyncAccount())
	require.Equal(t, expectedBalance, W1.GetBalance())
}

func TestTransfer_OverBalance(t *testing.T) {
	require.NoError(t, W0.Unlock(TESTPASS))

	ret, err := W1.TransferSync(W0.Address(), gas, amt)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrInsufficientFund.Error(), ret.Log)

	bal := W1.GetBalance()
	amt0 := new(big.Int).Sub(bal, gas)

	ret, err = W1.TransferSync(W0.Address(), amt0, gas)
	require.NoError(t, err) // not insufficient fund
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.Equal(t, txHash, txRet.Hash)
	require.NoError(t, W1.SyncAccount())
	require.Equal(t, big.NewInt(0), W1.GetBalance())
}
