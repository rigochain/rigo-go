package test

import (
	"bytes"
	"fmt"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestBulkTransfer(t *testing.T) {
	//for _, w := range wallets {
	//require.NoError(t, w.SyncAccount())
	//require.NoError(t, w.Unlock(TESTPASS))
	//}

	wg := sync.WaitGroup{}

	for i, w := range wallets {

		if bytes.Compare(w.Address(), validatorWallet.Address()) == 0 {
			continue
		}

		fmt.Println("index", i, "address", w.Address())

		require.NoError(t, w.SyncAccount())
		require.NoError(t, w.Unlock(TESTPASS))

		wg.Add(1)
		go bulkTransfer(t, &wg, w, 1000)
	}

	wg.Wait()
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, w *client.Wallet, cnt int) {

	fmt.Println("start address", w.Address())

	for i := 0; i < cnt; i++ {
		ret, err := w.TransferSync(types.RandAddress(), gas, amt)
		require.NoError(t, err)
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

		txHash := ret.Hash
		txRet, err := waitTrxResult(txHash, 10)
		require.NoError(t, err)

		require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
		require.Equal(t, txHash, txRet.Hash)
		require.Equal(t, gas, txRet.TxDetail.Gas)
		require.Equal(t, amt, txRet.TxDetail.Amount)

		expectedBalance := new(big.Int).Sub(w.GetBalance(), new(big.Int).Add(amt, gas))
		require.NoError(t, w.SyncAccount())
		require.Equal(t, expectedBalance, w.GetBalance(), w.Address())

		time.Sleep(1 * time.Millisecond)
	}
	wg.Done()

	fmt.Println("goodbye", w.Address())

}
