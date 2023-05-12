package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	rbytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/rand"
	"sync"
	"testing"
)

func TestTransfer_Bulk(t *testing.T) {

	wg := sync.WaitGroup{}

	var allAcctHelpers []*acctHelper
	senderCnt := 0
	for _, w := range wallets {
		if isValidatorWallet(w) {
			continue
		}

		require.NoError(t, w.SyncAccount(rweb3))

		acctTestObj := newAcctHelper(w)
		allAcctHelpers = append(allAcctHelpers, acctTestObj)

		//fmt.Println("TestBulkTransfer - used accounts:", w.Address(), w.GetNonce(), w.GetBalance())

		if senderCnt < 90 && w.GetBalance().Cmp(uint256.NewInt(1000000)) >= 0 {
			addSenderAcctHelper(w.Address().String(), acctTestObj)
			senderCnt++
		}
	}
	require.Greater(t, senderCnt, 1)

	// 최대 100 개 까지 계정 생성하여 리시버로 사용.
	// 100 개 이상이면 이미 있는 계정 사용.
	for i := len(allAcctHelpers); i < 100; i++ {
		newAcctTestObj := newAcctHelper(web3.NewWallet(TESTPASS))
		require.NoError(t, saveRandWallet(newAcctTestObj.w))
		allAcctHelpers = append(allAcctHelpers, newAcctTestObj)
	}

	for _, v := range senderAcctHelpers {
		wg.Add(1)
		go bulkTransfer(t, &wg, v, allAcctHelpers, 50) // 50 txs
	}

	wg.Wait()

	fmt.Printf("TestBulkTransfer - Check %v accounts ...\n", len(allAcctHelpers))

	for _, acctObj := range allAcctHelpers {
		//fmt.Println("\tCheck account", acctObj.w.Address())

		require.NoError(t, acctObj.w.SyncAccount(rweb3))
		require.Equal(t, acctObj.expectedBalance, acctObj.w.GetBalance(), acctObj.w.Address().String())
		require.Equal(t, acctObj.expectedNonce, acctObj.w.GetNonce(), acctObj.w.Address().String())
	}

	clearSenderAcctHelper()
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, senderAcctTestObj *acctHelper, receivers []*acctHelper, cnt int) {
	w := senderAcctTestObj.w
	require.NoError(t, w.Unlock(TESTPASS))

	//fmt.Printf("Begin of bulkTransfer - account: %v, balance: %v, nonce: %v\n", w.Address(), w.GetBalance(), w.GetNonce())

	subWg := sync.WaitGroup{}

	sub, err := web3.NewSubscriber(wsEndpoint)
	defer func() {
		sub.Stop()
	}()

	require.NoError(t, err)
	query := fmt.Sprintf("tm.event='Tx' AND tx.sender='%v'", w.Address())
	//fmt.Println("query", query)
	err = sub.Start(query, func(sub *web3.Subscriber, result []byte) {

		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)

		txHash, err := hex.DecodeString(event.Events["tx.hash"][0])
		require.NoError(t, err)

		addr, ok := senderAcctTestObj.txHashes[rbytes.HexBytes(txHash).String()]
		require.True(t, ok)
		require.Equal(t, w.Address(), addr)

		eventDataTx := event.Data.(tmtypes.EventDataTx)
		require.Equal(t, xerrors.ErrCodeSuccess, eventDataTx.TxResult.Result.Code)
		require.Equal(t, gas, uint256.NewInt(uint64(eventDataTx.TxResult.Result.GasUsed)))

		tx := &ctrlertypes.Trx{}
		err = tx.Decode(eventDataTx.Tx)
		require.NoError(t, err)

		//fmt.Println("bulkTransfer - event: ", rbytes.HexBytes(txHash), tx.From, tx.To, tx.Amount.Dec())

		subWg.Done()
	})
	require.NoError(t, err)

	maxAmt := new(uint256.Int).Div(senderAcctTestObj.originBalance, uint256.NewInt(uint64(cnt)))
	maxAmt = new(uint256.Int).Sub(maxAmt, gas)

	for i := 0; i < cnt; i++ {
		rn := rand.Intn(len(receivers))
		if bytes.Compare(receivers[rn].w.Address(), w.Address()) == 0 {
			rn = (rn + 1) % len(receivers)
		}

		racctState := receivers[rn]
		raddr := racctState.w.Address()

		randAmt := uint256.NewInt(1) //rbytes.RandU256IntN(maxAmt)
		if randAmt.Sign() == 0 {
			randAmt = uint256.NewInt(1)
		}
		//fmt.Printf("bulkTransfer - from: %v, to: %v, amount: %v\n", w.Address(), raddr, randAmt)
		needAmt := new(uint256.Int).Add(randAmt, gas)

		subWg.Add(1) // done in subscriber's callback

		ret, err := w.TransferSync(raddr, gas, randAmt, rweb3)
		require.NoError(t, err)
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log, w.GetNonce(), ret.Hash)

		// todo: this should be locked by mutex
		// record expected state of account
		senderAcctTestObj.addTxHashOfAddr(ret.Hash.String(), w.Address())
		senderAcctTestObj.addSpentGas(gas)
		senderAcctTestObj.subExpectedBalance(needAmt)
		senderAcctTestObj.addExpectedNonce()

		racctState.addExpectedBalance(randAmt)

		fmt.Printf("Send Tx [txHash: %v, from: %v, to: %v, nonce: %v, amt: %v]\n", ret.Hash, w.Address(), racctState.w.Address(), w.GetNonce(), randAmt)

		w.AddNonce()
	}

	subWg.Wait()

	wg.Done()

	//fmt.Printf("End of bulkTransfer - account: %v, balance: %v, nonce: %v\n", w.Address(), w.GetBalance(), w.GetNonce())
}

func TestTransfer_OverBalance(t *testing.T) {
	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W1.SyncBalance(rweb3))
	require.NoError(t, W0.Unlock(TESTPASS))

	testObj0 := newAcctHelper(W0)
	testObj1 := newAcctHelper(W1)

	overAmt := W0.GetBalance() // gas is not included

	ret, err := W0.TransferSync(W1.Address(), gas, overAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	//require.Equal(t, xerrors.ErrCheckTx.Wrap(xerrors.ErrInsufficientFund).Error(), ret.Log)

	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W1.SyncBalance(rweb3))

	require.Equal(t, testObj0.originBalance, W0.GetBalance())
	require.Equal(t, testObj1.originBalance, W1.GetBalance())
	require.Equal(t, testObj0.originNonce, W0.GetNonce())

	overAmt = new(uint256.Int).Add(new(uint256.Int).Sub(W0.GetBalance(), gas), uint256.NewInt(1)) // amt - gas + 1
	ret, err = W0.TransferSync(W1.Address(), gas, overAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	//require.Equal(t, xerrors.ErrCheckTx.Wrap(xerrors.ErrInsufficientFund).Error(), ret.Log)

	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W1.SyncBalance(rweb3))

	require.Equal(t, testObj0.originBalance, W0.GetBalance())
	require.Equal(t, testObj1.originBalance, W1.GetBalance())
	require.Equal(t, testObj0.originNonce, W0.GetNonce())

}

func TestTransfer_WrongAddr(t *testing.T) {
	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W0.Unlock(TESTPASS))
	require.NotEqual(t, uint256.NewInt(0).String(), W0.GetBalance().String())

	tmpAmt := new(uint256.Int).Div(W0.GetBalance(), uint256.NewInt(2))
	ret, err := W0.TransferSync(nil, gas, tmpAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Code)

	ret, err = W0.TransferSync([]byte{0x00}, gas, tmpAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Code)
}
