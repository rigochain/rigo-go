package test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3/vm"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
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
	testDeploy(t)
	testPayable(t)
}

func testDeploy(t *testing.T) {
	rweb3 := randRigoWeb3()

	creator := validatorWallets[0]
	require.NoError(t, creator.SyncAccount(rweb3))
	require.NoError(t, creator.Unlock(defaultRpcNode.Pass), string(defaultRpcNode.Pass))

	contract, err := vm.NewEVMContract("./erc20_test_contract.json")
	require.NoError(t, err)

	ret, err := contract.Exec("", []interface{}{"RigoToken", "RGT"},
		creator, creator.GetNonce(), gas, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	//require.NotNil(t, ret.Data)
	//require.Equal(t, 20, len(ret.Data))

	txRet, err := waitTrxResult(ret.Hash, 15, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log)
	require.NotNil(t, txRet.TxResult.Data)

	contract.SetAddress(txRet.TxResult.Data)
	evmContract = contract
}

func testQuery(t *testing.T) {
	rweb3 := randRigoWeb3()

	sender := randCommonWallet()
	require.NoError(t, sender.SyncAccount(rweb3))
	ret, err := evmContract.Call("name", nil, sender.Address(), 0, rweb3)
	require.NoError(t, err)
	require.Equal(t, "RigoToken", ret[0])
}

func testPayable(t *testing.T) {
	rweb3 := randRigoWeb3()

	sender := randCommonWallet()
	require.NoError(t, sender.Unlock(defaultRpcNode.Pass), string(defaultRpcNode.Pass))
	require.NoError(t, sender.SyncAccount(rweb3))

	contAcct, err := rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	require.Equal(t, "0", contAcct.Balance.Dec())

	//fmt.Println("initial", "sender", sender.Address(), "balance", sender.GetBalance())
	//fmt.Println("initial", "contAcct", contAcct.Address, "balance", contAcct.GetBalance())

	//
	// Transfer
	//
	randAmt := bytes.RandU256IntN(sender.GetBalance())
	_ = randAmt.Sub(randAmt, gas)
	_ = randAmt.Sub(randAmt, gas)

	ret, err := sender.TransferSync(evmContract.GetAddress(), gas, randAmt, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	txRet, err := waitTrxResult(ret.Hash, 15, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)

	expectedAmt := new(uint256.Int).Sub(sender.GetBalance(), uint256.NewInt(uint64(txRet.TxResult.GasUsed)))
	_ = expectedAmt.Sub(expectedAmt, randAmt)
	require.NotEqual(t, sender.GetBalance(), expectedAmt)
	require.NoError(t, sender.SyncAccount(rweb3))
	require.Equal(t, expectedAmt, sender.GetBalance())

	contAcct, err = rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	require.Equal(t, randAmt, contAcct.Balance)

	//fmt.Println("after transfer", "sender", sender.Address(), "balance", sender.GetBalance())
	//fmt.Println("after transfer", "contAcct", contAcct.Address, "balance", contAcct.GetBalance())

	//
	// payable function giveMeAsset
	//

	refundAmt := bytes.RandU256IntN(randAmt)
	ret, err = evmContract.Exec("giveMeAsset", []interface{}{refundAmt.ToBig()}, sender, sender.GetNonce(), gas, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	txRet, err = waitTrxResult(ret.Hash, 15, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)

	expectedAmt = new(uint256.Int).Add(sender.GetBalance(), refundAmt)
	_ = expectedAmt.Sub(expectedAmt, uint256.NewInt(uint64(txRet.TxResult.GasUsed)))
	require.NoError(t, sender.SyncAccount(rweb3))
	require.Equal(t, expectedAmt, sender.GetBalance())

	expectedAmt = new(uint256.Int).Sub(contAcct.GetBalance(), refundAmt)
	contAcct, err = rweb3.GetAccount(evmContract.GetAddress())
	require.NoError(t, err)
	require.Equal(t, expectedAmt, contAcct.Balance)

	fmt.Println("after refund", "sender", sender.Address(), "balance", sender.GetBalance())
	fmt.Println("after refund", "contAcct", contAcct.Address, "balance", contAcct.GetBalance())
}
