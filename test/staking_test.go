package test

import (
	"bytes"
	"fmt"
	rtypes0 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/big"
	"testing"
)

func TestQueryValidators(t *testing.T) {
	ret, err := validators(1)

	require.NoError(t, err)
	require.Equal(t, len(validatorWallets), len(ret.Validators))
	for _, val := range ret.Validators {
		require.True(t, isValidator(rtypes0.Address(val.Address)))
	}
}

func TestInvalidStakeAmount(t *testing.T) {
	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))
	require.NoError(t, w.Unlock(TESTPASS))

	// too small
	stakeAmt, ok := new(big.Int).SetString("1111", 10)
	require.True(t, ok)

	ret, err := w.StakingSync(w.Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	t.Log(ret.Log)

	// not multiple
	stakeAmt, ok = new(big.Int).SetString("1000000000000000001", 10)
	require.True(t, ok)

	ret, err = w.StakingSync(w.Address(), gas, stakeAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	t.Log(ret.Log)
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
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	query := fmt.Sprintf("tm.event='tx' AND tx.hash='%v'", ret.Hash)
	wg, err := waitEvent(query, func(evt *coretypes.ResultEvent, err error) bool {
		require.NoError(t, err)
		evtTx, ok := evt.Data.(*tmtypes.EventDataTx)
		require.True(t, ok)

		require.Equal(t, xerrors.ErrCodeSuccess, evtTx.Result.Code)
		require.Equal(t, gas.Int64(), evtTx.Result.GasUsed)

		return true // done and stop subscriber
	})
	wg.Wait()

	// check stakes
	found := false
	stakes, err := rweb3.GetStakes(w.Address())
	require.True(t, len(stakes) > 0)
	for _, s0 := range stakes {
		if bytes.Compare(s0.TxHash, ret.Hash) == 0 {
			require.Equal(t, stakeAmt, s0.Amount)
			found = true
		}
	}
	require.True(t, found)
	require.NoError(t, err)

	addValidatorWallet(w)
}

func TestDelegating(t *testing.T) {

	w := randCommonWallet()
	require.NoError(t, w.SyncAccount(rweb3))

	bal := w.GetBalance()
	require.True(t, bal.Cmp(big.NewInt(0)) > 0)

	stakeAmt := big.NewInt(1000000000000000000) //rbytes.RandBigIntN(bal)

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
