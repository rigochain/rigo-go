package test

import (
	"encoding/hex"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestStaking(t *testing.T) {
	addr, err := hex.DecodeString("4754BD89C3E067A5405F0FCF1158EB1A252362D8")
	require.NoError(t, err)

	acctKey := account.ToAcctKey(addr)
	w, ok := walletsMap[acctKey]
	require.True(t, ok)
	require.NoError(t, w.SyncAccount())

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt, ok := new(big.Int).SetString("15000000000000000000", 10)
	require.True(t, ok)

	require.NoError(t, w.Unlock(TESTPASS))
	ret, err := w.StakingSync(w.Address(), stakeAmt, gas)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	txHash := ret.Hash

	txRet, err := waitTrxResult(txHash, 10)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
	require.Equal(t, txHash, txRet.Hash)
	//require.Equal(t, trxs.TRX_TRANSFER, txRet.TxDetail.Payload.Type())
	require.Equal(t, gas, txRet.TxDetail.Gas)
	require.Equal(t, stakeAmt, txRet.TxDetail.Amount)

	// todo: query stake amount and check that the response is correct
}
