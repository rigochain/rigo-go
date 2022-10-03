package test

import (
	"fmt"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestBulkTransfer(t *testing.T) {
	for _, w := range wallets {
		require.NoError(t, w.SyncAccount())
		require.NoError(t, w.Unlock(TESTPASS))
	}

	wg := sync.WaitGroup{}

	for _, w := range wallets {
		wg.Add(1)
		go bulkTransfer(t, &wg, w, 10000)
	}

	wg.Wait()
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, w *client.Wallet, cnt int) {

	fmt.Println("address", w.Address())

	for i := 0; i < cnt; i++ {
		ret, err := w.TransferSync(libs.RandAddress(), amt, gas)
		require.NoError(t, err)
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

		//txHash := ret.Hash
		//txRet, err := waitTrxResult(txHash, 10)
		//require.NoError(t, err)
		//
		//require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
		//require.Equal(t, txHash, txRet.Hash)
		////require.Equal(t, trxs.TRX_TRANSFER, txRet.TxDetail.Payload.Type())
		//require.Equal(t, gas, txRet.TxDetail.Gas)
		//require.Equal(t, amt, txRet.TxDetail.Amount)
		//
		//expectedBalance := new(big.Int).Sub(w.GetBalance(), new(big.Int).Add(amt, gas))
		//require.NoError(t, w.SyncAccount())
		//require.Equal(t, expectedBalance, w.GetBalance())
		//
		////expectedBalance = new(big.Int).Add(W1.GetBalance(), amt)
		////require.NoError(t, W1.SyncAccount())
		////require.Equal(t, expectedBalance, W1.GetBalance())

		time.Sleep(1 * time.Millisecond)
	}
	wg.Done()
}
