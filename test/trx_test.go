package test

import (
	"github.com/kysee/arcanus/libs/client/rpc"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestNonce(t *testing.T) {
	require.NoError(t, W0.SyncAccount())
	require.NoError(t, W0.Unlock(TESTPASS))
	ret, err := W0.TransferSync(W1.Address(), amt, gas)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)
	txHash := ret.Hash

	_, err = waitTrxResult(txHash, 10)
	require.NoError(t, err)

	// do not add/sync nonce

	ret, err = W0.TransferSync(W1.Address(), amt, gas)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrInvalidNonce.Error(), ret.Log)

	// return fund
	require.NoError(t, W1.SyncAccount())

	bal := W1.GetBalance()
	amt0 := new(big.Int).Sub(bal, gas)

	ret, err = W1.TransferSync(W0.Address(), amt0, gas)
	require.NoError(t, err) // not insufficient fund
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)
	txHash = ret.Hash

	_, err = waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.NoError(t, W1.SyncAccount())
	require.Equal(t, big.NewInt(0), W1.GetBalance())
}

func waitTrxResult(txhash []byte, maxTimes int) (*rpc.TrxResult, error) {
	for i := 0; i < maxTimes; i++ {
		time.Sleep(1500 * time.Millisecond)

		// todo: check why it takes more than 10 secs to fetch a transaction

		txRet, err := rpc.GetTransaction(txhash)
		if err != nil && !strings.Contains(err.Error(), ") not found") {
			return nil, err
		} else if err == nil {
			return txRet, nil
		}
	}
	return nil, xerrors.New("timeout")
}
