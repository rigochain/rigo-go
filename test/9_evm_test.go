package test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/libs/web3/vm"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"sync"
	"testing"
)

var (
	evmContract *vm.EVMContract
	creator     *web3.Wallet
)

func TestERC20_Deploy(t *testing.T) {
	// deploy
	testDeploy(t)
	testQuery(t)
}

func TestERC20_Payable(t *testing.T) {
	testDeploy(t)
	testPayable(t)
}

func TestERC20_Event(t *testing.T) {
	testDeploy(t)
	testEvents(t)
}

func testDeploy(t *testing.T) {
	rweb3 := randRigoWeb3()

	creator = randCommonWallet()
	require.NoError(t, creator.Unlock(defaultRpcNode.Pass), string(defaultRpcNode.Pass))

	require.NoError(t, creator.SyncAccount(rweb3))
	beforeBalance0 := creator.GetBalance().Clone()

	contract, err := vm.NewEVMContract("./erc20_test_contract.json")
	require.NoError(t, err)

	// insufficient gas
	ret, err := contract.Exec("", []interface{}{"RigoToken", "RGT"},
		creator, creator.GetNonce(), baseFee, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	// check balance - not changed
	require.NoError(t, creator.SyncAccount(rweb3))
	beforeBalance1 := creator.GetBalance().Clone()
	require.Equal(t, beforeBalance0.Dec(), beforeBalance1.Dec())

	// sufficient gas
	ret, err = contract.Exec("", []interface{}{"RigoToken", "RGT"},
		creator, creator.GetNonce(), limitFee, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)
	require.NotNil(t, ret.Data)
	require.Equal(t, 20, len(ret.Data))

	txRet, err := waitTrxResult(ret.Hash, 30, rweb3)
	require.NoError(t, err, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log)
	require.NotNil(t, txRet.TxResult.Data)

	contract.SetAddress(txRet.TxResult.Data)
	evmContract = contract

	require.NoError(t, creator.SyncAccount(rweb3))
	afterBalance := creator.GetBalance().Clone()

	// check balance - changed by gas
	usedGas := new(uint256.Int).Sub(beforeBalance1, afterBalance).Uint64()
	require.Equal(t, uint64(txRet.TxResult.GasUsed), usedGas)
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
	_ = randAmt.Sub(randAmt, baseFee)

	ret, err := sender.TransferSync(evmContract.GetAddress(), baseFee, randAmt, rweb3)
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
	ret, err = evmContract.Exec("giveMeAsset", []interface{}{refundAmt.ToBig()}, sender, sender.GetNonce(), limitFee, uint256.NewInt(0), rweb3)
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
	require.Equal(t, expectedAmt, contAcct.GetBalance())

	fmt.Println("after refund", "sender", sender.Address(), "balance", sender.GetBalance())
	fmt.Println("after refund", "contAcct", contAcct.Address, "balance", contAcct.GetBalance())
}

func testEvents(t *testing.T) {
	rweb3 := randRigoWeb3()

	require.NoError(t, creator.Unlock(defaultRpcNode.Pass), string(defaultRpcNode.Pass))
	require.NoError(t, creator.SyncAccount(rweb3))

	// subcribe events
	subWg := &sync.WaitGroup{}
	sub, err := web3.NewSubscriber(defaultRpcNode.WSEnd)
	defer func() {
		sub.Stop()
	}()
	require.NoError(t, err)
	query := fmt.Sprintf("tx.type='contract' AND evm.topic.0='ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef'")
	err = sub.Start(query, func(sub *web3.Subscriber, result []byte) {
		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)

		subWg.Done()
	})
	require.NoError(t, err)

	// broadcast tx
	subWg.Add(1)

	rAddr := types.RandAddress()
	// Transfer Event sig: ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	ret, err := evmContract.Exec("transfer", []interface{}{rAddr.Array20(), uint256.NewInt(100).ToBig()}, creator, creator.GetNonce(), limitFee, uint256.NewInt(0), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	txRet, err := waitTrxResult(ret.Hash, 15, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)

	subWg.Wait()

}
