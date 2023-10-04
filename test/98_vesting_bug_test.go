package test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3/vm"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBalanceBug(t *testing.T) {
	rweb3 := randRigoWeb3()

	deployer := randCommonWallet() // web3.NewWallet(defaultRpcNode.Pass)
	require.NoError(t, deployer.Unlock(defaultRpcNode.Pass), string(defaultRpcNode.Pass))
	require.NoError(t, deployer.SyncAccount(rweb3))
	fmt.Println("deployer address", deployer.Address(), "balance", deployer.GetBalance().Dec(), "nonce", deployer.GetNonce())

	contract, err := vm.NewEVMContract("./vesting_test.json")
	require.NoError(t, err)

	// deploy
	ret, err := contract.ExecCommit("", []interface{}{},
		deployer, deployer.GetNonce(), contractGas, defGasPrice, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.CheckTx.Code, ret.CheckTx.Log)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.DeliverTx.Code, ret.DeliverTx.Log)
	require.NotNil(t, contract.GetAddress())

	contAddr := contract.GetAddress()
	fmt.Println("contract address", contAddr)

	// get contract's balance
	require.NoError(t, deployer.SyncAccount(rweb3))
	retCall, err := contract.Call("userBalance", []interface{}{contAddr.Array20()}, contAddr, 0, rweb3)
	require.NoError(t, err)

	fmt.Println("userBalance returns", retCall)

	//fmt.Println("return value", new(uint256.Int).SetBytes(retCall).Dec())

	// transfer to contract
	require.NoError(t, deployer.SyncAccount(rweb3))
	_amt := new(uint256.Int).Div(deployer.GetBalance(), uint256.NewInt(2))
	ret, err = deployer.TransferCommit(contAddr, defGas, defGasPrice, _amt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.CheckTx.Code, ret.CheckTx.Log)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.DeliverTx.Code, ret.DeliverTx.Log)

	fmt.Println("transfer amount", _amt.Dec())

	// get contract's balance
	require.NoError(t, deployer.SyncAccount(rweb3))
	retCall, err = contract.Call("userBalance", []interface{}{contAddr.Array20()}, contAddr, 0, rweb3)
	require.NoError(t, err)
	fmt.Println("userBalance returns", retCall)
	//fmt.Println("return value", new(uint256.Int).SetBytes(ret.DeliverTx.Data).Dec())

}
