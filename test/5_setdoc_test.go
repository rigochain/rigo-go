package test

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestSetDoc(t *testing.T) {
	rweb3 := randRigoWeb3()

	w := randCommonWallet()
	require.NoError(t, w.Unlock(defaultRpcNode.Pass))
	require.NoError(t, w.SyncAccount(rweb3))

	oriBalance := w.GetBalance().Clone()
	name := "test account"
	url := "https://www.my.site/doc"

	ret, err := w.SetDocSync(name, url, smallGas, defGasPrice, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)

	ret, err = w.SetDocSync(name, url, defGas, defGasPrice, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code)

	txRet, xerr := waitTrxResult(ret.Hash, 30, rweb3)
	require.NoError(t, xerr)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log)

	expectedBalance := new(uint256.Int).Sub(oriBalance, gasToFee(uint64(txRet.TxResult.GasUsed), defGasPrice))
	require.NoError(t, w.SyncAccount(rweb3))
	require.Equal(t, expectedBalance.Dec(), w.GetBalance().Dec())
	require.Equal(t, name, w.GetAccount().Name)
	require.Equal(t, url, w.GetAccount().DocURL)

	tooLongName := rand.Str(types.MAX_ACCT_NAME + 1)
	ret, err = w.SetDocSync(tooLongName, url, defGas, defGasPrice, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

}
