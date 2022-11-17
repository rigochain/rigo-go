package test

import (
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

var (
	WALKEYDIR  = ".tmp/walkeys"
	TESTPASS   = []byte("1")
	wallets    []*client.Wallet
	walletsMap map[account.AcctKey]*client.Wallet
	W0         *client.Wallet
	W1         *client.Wallet
	amt        = libs.RandBigIntN(big.NewInt(1000))
	gas        = big.NewInt(10)
)

func init() {

	files, err := os.ReadDir(WALKEYDIR)
	if err != nil {
		panic(err)
	}

	walletsMap = make(map[account.AcctKey]*client.Wallet)

	for _, file := range files {
		if !file.IsDir() {
			if w, err := client.OpenWallet(
				libs.NewFileReader(filepath.Join(WALKEYDIR, file.Name()))); err != nil {
				panic(err)
			} else {
				wallets = append(wallets, w)

				acctKey := account.ToAcctKey(w.Address())
				walletsMap[acctKey] = w
			}
		}
	}
	W0 = wallets[0]
	W1 = wallets[1]
}

func TestSyncWallet(t *testing.T) {
	for _, w := range wallets {
		require.NoError(t, w.SyncAccount())
	}
}

func TestTransfer(t *testing.T) {

	ret, err := W0.TransferSync(W1.Address(), amt, gas)
	require.Error(t, err) // unlock

	require.NoError(t, W0.Unlock(TESTPASS))

	ret, err = W0.TransferSync(W1.Address(), amt, gas)
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

	ret, err := W1.TransferSync(W0.Address(), amt, gas)
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
