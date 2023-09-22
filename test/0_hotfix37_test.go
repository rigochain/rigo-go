package test

import (
	"fmt"
	"github.com/holiman/uint256"
	types3 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"testing"
)

// when validator does not exist in ledger and tx fee should be given to the validator.
// a panic is generated in previous version.
func TestTransfer0(t *testing.T) {
	rweb3 := randRigoWeb3()

	sender := randCommonWallet()
	require.NoError(t, sender.Unlock([]byte("1111")))
	require.NoError(t, sender.SyncAccount(rweb3))

	receiver := randCommonWallet()

	txRet, err := sender.TransferCommit(receiver.Address(), defGas, defGasPrice, types3.ToFons(1), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.CheckTx.Code, txRet.CheckTx.Log)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.DeliverTx.Code, txRet.DeliverTx.Log)

	// For next test
	// transfer asset to validator
	require.NoError(t, sender.SyncAccount(rweb3))
	_amt := new(uint256.Int).Div(sender.GetBalance(), uint256.NewInt(2))
	fmt.Println("wallet amount", sender.GetBalance().Dec())
	fmt.Println("transfer amount", _amt.Dec())

	val0 := validatorWallets[0]
	txRet, err = sender.TransferCommit(val0.Address(), defGas, defGasPrice, _amt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.CheckTx.Code, txRet.CheckTx.Log)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.DeliverTx.Code, txRet.DeliverTx.Log)

	require.NoError(t, val0.SyncAccount(rweb3))
	fmt.Println("validator", val0.Address(), "balance", val0.GetBalance().Dec())
}

// Disable test case
// the validator has already over minPower.

//func TestStaking2GenesisValidator(t *testing.T) {
//	rweb3 := randRigoWeb3()
//	govParams, err := rweb3.GetGovParams()
//	require.NoError(t, err)
//
//	valWal := validatorWallets[0]
//	require.NoError(t, valWal.SyncAccount(rweb3))
//	require.NoError(t, valWal.Unlock(defaultRpcNode.Pass))
//
//	valStakes0, err := rweb3.GetDelegatee(valWal.Address())
//	require.NoError(t, err)
//
//	fmt.Println("valStake0.SelfAmount", valStakes0.SelfPower)
//
//	amtStake := uint256.NewInt(1000000000000000000)
//	ret, err := valWal.StakingCommit(valWal.Address(), defGas, defGasPrice, amtStake, rweb3)
//	require.NoError(t, err)
//	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.CheckTx.Code)
//	require.Contains(t, ret.CheckTx.Log, "too small stake to become validator", ret.CheckTx.Log)
//
//	amtStake = new(uint256.Int).Sub(govParams.MinValidatorStake(), types.PowerToAmount(valStakes0.SelfPower))
//	ret, err = valWal.StakingCommit(valWal.Address(), defGas, defGasPrice, amtStake, rweb3)
//	require.NoError(t, err)
//	require.Equal(t, xerrors.ErrCodeSuccess, ret.DeliverTx.Code, ret.DeliverTx.Log)
//
//	stakes, err := rweb3.GetStakes(valWal.Address())
//	require.NoError(t, err)
//	require.Equal(t, 2, len(stakes), stakes)
//
//	found := false
//	for _, s := range stakes {
//		if bytes.Compare(ret.Hash, s.TxHash) == 0 {
//			found = true
//			break
//		}
//	}
//	require.True(t, found)
//
//	valStakes1, err := rweb3.GetDelegatee(valWal.Address())
//	require.NoError(t, err)
//	require.Equal(t,
//		new(uint256.Int).Add(types.PowerToAmount(valStakes0.TotalPower), amtStake),
//		types.PowerToAmount(valStakes1.TotalPower))
//	require.Equal(t, valStakes1.TotalPower,
//		valStakes1.SumPower())
//
//	//fmt.Println("valStakes1.SelfAmount", valStakes1.SelfAmount.Dec())
//
//}
