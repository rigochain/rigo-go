package test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3/vm"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

var (
	evmContract *vm.EVMContract
)

func TestERC20_Deploy(t *testing.T) {
	// deploy
	testDeploy(t)
}

func TestERC20_Query(t *testing.T) {
	testQuery(t)
}

func TestERC20_Payable(t *testing.T) {
	testPayable(t)
}

func testDeploy(t *testing.T) {
	creator := validatorWallets[0]
	require.NoError(t, creator.SyncAccount(rweb3))
	require.NoError(t, creator.Unlock(TESTPASS))

	contract, err := vm.NewEVMContract("./erc20_test_contract.json")
	require.NoError(t, err)

	ret, err := contract.Exec("", []interface{}{"RigoToken", "RGT"},
		creator, creator.GetNonce(), gas, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	require.NotNil(t, ret.Data)
	require.Equal(t, 20, len(ret.Data))

	waitTrxResult(ret.Hash, 15)

	evmContract = contract
}

func testQuery(t *testing.T) {
	sender := randCommonWallet()
	require.NoError(t, sender.SyncAccount(rweb3))
	ret, err := evmContract.Call("name", nil, sender.Address(), 0, rweb3)
	require.NoError(t, err)
	require.Equal(t, "RigoToken", ret[0])
}

func testPayable(t *testing.T) {
	sender := randCommonWallet()
	require.NoError(t, sender.Unlock(TESTPASS))
	require.NoError(t, sender.SyncAccount(rweb3))
	fmt.Println("sender", sender.Address(), sender.GetBalance().Dec())
	contAcct, err := rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	fmt.Println("receiver", contAcct.Address, contAcct.Balance.Dec())

	ret, err := sender.TransferSync(evmContract.GetAddress(), gas, uint256.NewInt(100), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	waitTrxResult(ret.Hash, 15)

	require.NoError(t, sender.SyncAccount(rweb3))
	fmt.Println("sender", sender.Address(), sender.GetBalance().Dec())

	contAcct, err = rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	fmt.Println("receiver", contAcct.Address, contAcct.Balance.Dec())

	ret, err = evmContract.Exec("giveMeAsset", []interface{}{big.NewInt(10)}, sender, sender.GetNonce(), gas, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	waitTrxResult(ret.Hash, 15)

	require.NoError(t, sender.SyncAccount(rweb3))
	fmt.Println("sender", sender.Address(), sender.GetBalance().Dec())

	contAcct, err = rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	fmt.Println("receiver", contAcct.Address, contAcct.Balance.Dec())

}
