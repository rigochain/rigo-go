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
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTransfer_Bulk(t *testing.T) {
	rweb3 := randRigoWeb3()

	wg := sync.WaitGroup{}

	var allAcctObjs []*acctObj
	senderCnt := 0
	for _, w := range wallets {
		if isValidatorWallet(w) {
			continue
		}

		require.NoError(t, w.SyncAccount(rweb3))

		acctTestObj := newAcctObj(w)
		allAcctObjs = append(allAcctObjs, acctTestObj)
		//fmt.Println("TestBulkTransfer - used accounts:", w.Address(), w.GetNonce(), w.GetBalance())

		if senderCnt < 90 && w.GetBalance().Cmp(uint256.NewInt(1000000)) >= 0 {
			addSenderAcctHelper(w.Address().String(), acctTestObj)
			senderCnt++
		}
	}
	require.Greater(t, senderCnt, 1)

	//// 최대 100 개 까지 계정 생성하여 리시버로 사용.
	//// 100 개 이상이면 이미 있는 계정 사용.
	for i := len(allAcctObjs); i < 100; i++ {
		newAcctTestObj := newAcctObj(web3.NewWallet(defaultRpcNode.Pass))
		require.NoError(t, saveWallet(newAcctTestObj.w))
		allAcctObjs = append(allAcctObjs, newAcctTestObj)
	}

	for _, v := range senderAcctObjs {
		wg.Add(1)
		go bulkTransfer(t, &wg, v, allAcctObjs, 500) // 100 txs per sender
	}

	fmt.Printf("TestBulkTransfer - Sender accounts: %d, Receiver accounts: %d\n", len(senderAcctObjs), len(allAcctObjs))

	wg.Wait()

	fmt.Printf("TestBulkTransfer - Check %v accounts ...\n", len(allAcctObjs))

	sentTxsCnt, retTxsCnt := 0, 0
	for _, acctObj := range allAcctObjs {
		//fmt.Println("\tCheck account", acctObj.w.Address())

		//for k, _ := range acctObj.txHashes {
		//	hash, err := hex.DecodeString(k)
		//	require.NoError(t, err)
		//
		//	ret, err := waitTrxResult(hash, 50)
		//	require.NoError(t, err, k)
		//	require.Equal(t, xerrors.ErrCodeSuccess, ret.TxResult.Code,
		//		fmt.Sprintf("error: %v, address: %v, txhash: %v", ret.TxResult.Log, acctObj.w.Address(), ret.Hash))
		//}

		require.NoError(t, acctObj.w.SyncAccount(rweb3))
		require.Equal(t, acctObj.expectedBalance, acctObj.w.GetBalance(), acctObj.w.Address().String())
		require.Equal(t, acctObj.expectedNonce, acctObj.w.GetNonce(), acctObj.w.Address().String())

		sentTxsCnt += acctObj.sentTxsCnt
		retTxsCnt += acctObj.retTxsCnt
	}

	fmt.Printf("TestBulkTransfer - senders: %d, sent txs: %d, result txs: %d\n", senderCnt, sentTxsCnt, retTxsCnt)

	clearSenderAcctHelper()
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, senderAcctObj *acctObj, receivers []*acctObj, cnt int) {
	w := senderAcctObj.w
	require.NoError(t, w.Unlock(defaultRpcNode.Pass))

	rpcNode := randPeer()
	fmt.Printf("Begin of bulkTransfer - account: %v, balance: %v, nonce: %v, rpcPeerIdx: %v\n", w.Address(), w.GetBalance(), w.GetNonce(), rpcNode.PeerID)
	_rweb3 := web3.NewRigoWeb3(web3.NewHttpProvider(rpcNode.RPCURL))

	subWg := &sync.WaitGroup{}
	sub, err := web3.NewSubscriber(rpcNode.WSEnd)
	defer func() {
		sub.Stop()
	}()
	require.NoError(t, err)
	query := fmt.Sprintf("tm.event='Tx' AND tx.sender='%v'", w.Address())
	err = sub.Start(query, func(sub *web3.Subscriber, result []byte) {

		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)

		txHash, err := hex.DecodeString(event.Events["tx.hash"][0])
		require.NoError(t, err)

		addr := senderAcctObj.getAddrOfTxHash(txHash)
		require.Equal(t, w.Address(), addr)

		eventDataTx := event.Data.(tmtypes.EventDataTx)
		require.Equal(t, xerrors.ErrCodeSuccess, eventDataTx.TxResult.Result.Code)
		require.Equal(t, gas10, uint256.NewInt(uint64(eventDataTx.TxResult.Result.GasUsed)))

		tx := &ctrlertypes.Trx{}
		err = tx.Decode(eventDataTx.Tx)
		require.NoError(t, err)

		senderAcctObj.retTxsCnt++

		subWg.Done()

	})
	require.NoError(t, err)

	//checkTxRoutine := func(txhash []byte) {
	//	retTx, err := waitTrxResult(txhash, 10, _rweb3)
	//	require.NoError(t, err)
	//	require.Equal(t, xerrors.ErrCodeSuccess, retTx.TxResult.Code, retTx.TxResult.Log)
	//	subWg.Done()
	//}

	maxAmt := new(uint256.Int).Div(senderAcctObj.originBalance, uint256.NewInt(uint64(cnt)))
	maxAmt = new(uint256.Int).Sub(maxAmt, gas10)

	for i := 0; i < cnt; i++ {
		rn := rand.Intn(len(receivers))
		if bytes.Compare(receivers[rn].w.Address(), w.Address()) == 0 {
			rn = (rn + 1) % len(receivers)
		}

		racctState := receivers[rn]
		raddr := racctState.w.Address()

		randAmt := rbytes.RandU256IntN(maxAmt)
		if randAmt.Sign() == 0 {
			randAmt = uint256.NewInt(1)
		}
		//fmt.Printf("bulkTransfer - from: %v, to: %v, amount: %v\n", w.Address(), raddr, randAmt)
		needAmt := new(uint256.Int).Add(randAmt, gas10)

		subWg.Add(1)

		ret, err := w.TransferSync(raddr, gas10, randAmt, _rweb3)

		if err != nil && strings.Contains(err.Error(), "mempool is full") {
			subWg.Done()
			time.Sleep(time.Millisecond * 1000)

			i--
			continue
		}
		require.NoError(t, err)

		if ret.Code == xerrors.ErrCodeCheckTx &&
			strings.Contains(ret.Log, "invalid nonce") {

			subWg.Done()
			require.NoError(t, w.SyncAccount(_rweb3))

			i--
			continue
		}
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log, w.GetNonce(), ret.Hash)

		//checkTxRoutine(ret.Hash)

		senderAcctObj.addTxHashOfAddr(ret.Hash, w.Address())
		senderAcctObj.addSpentGas(gas10)
		senderAcctObj.subExpectedBalance(needAmt)
		senderAcctObj.addExpectedNonce()
		racctState.addExpectedBalance(randAmt)

		//fmt.Printf("Send Tx [txHash: %v, from: %v, to: %v, nonce: %v, amt: %v]\n", ret.Hash, w.Address(), racctState.w.Address(), w.GetNonce(), randAmt)

		w.AddNonce()

		senderAcctObj.sentTxsCnt++
	}
	//fmt.Println(senderAcctObj.w.Address(), "sent", senderAcctObj.sentTxsCnt, "ret", senderAcctObj.retTxsCnt)
	subWg.Wait()
	//fmt.Println(senderAcctObj.w.Address(), "sent", senderAcctObj.sentTxsCnt, "ret", senderAcctObj.retTxsCnt)

	wg.Done()

	//fmt.Printf("End of bulkTransfer - account: %v, balance: %v, nonce: %v\n", w.Address(), w.GetBalance(), w.GetNonce())
}

func TestTransfer_OverBalance(t *testing.T) {
	rweb3 := randRigoWeb3()

	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W1.SyncBalance(rweb3))
	require.NoError(t, W0.Unlock(defaultRpcNode.Pass))

	testObj0 := newAcctObj(W0)
	testObj1 := newAcctObj(W1)

	overAmt := W0.GetBalance() // gas10 is not included

	ret, err := W0.TransferSync(W1.Address(), gas10, overAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	//require.Equal(t, xerrors.ErrCheckTx.Wrap(xerrors.ErrInsufficientFund).Error(), ret.Log)

	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W1.SyncBalance(rweb3))

	require.Equal(t, testObj0.originBalance, W0.GetBalance())
	require.Equal(t, testObj1.originBalance, W1.GetBalance())
	require.Equal(t, testObj0.originNonce, W0.GetNonce())

	overAmt = new(uint256.Int).Add(new(uint256.Int).Sub(W0.GetBalance(), gas10), uint256.NewInt(1)) // amt - gas10 + 1
	ret, err = W0.TransferSync(W1.Address(), gas10, overAmt, rweb3)
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
	rweb3 := randRigoWeb3()

	require.NoError(t, W0.SyncBalance(rweb3))
	require.NoError(t, W0.Unlock(defaultRpcNode.Pass))
	require.NotEqual(t, uint256.NewInt(0).String(), W0.GetBalance().String())

	tmpAmt := new(uint256.Int).Div(W0.GetBalance(), uint256.NewInt(2))
	ret, err := W0.TransferSync(nil, gas10, tmpAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Code)

	ret, err = W0.TransferSync([]byte{0x00}, gas10, tmpAmt, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code, ret.Code)
}
