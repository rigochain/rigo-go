package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/libs/client"
	"github.com/rigochain/rigo-go/types"
	rbytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/big"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
)

type accountTestObj struct {
	w *client.Wallet

	originBalance *big.Int
	originNonce   uint64

	txHashes        map[string]types.Address
	spentGas        *big.Int
	expectedBalance *big.Int
	expectedNonce   uint64
}

var senderAccountTestObjMap = make(map[string]*accountTestObj)
var gmtx = &sync.Mutex{}

func newAccountTestObj(w *client.Wallet) *accountTestObj {
	return &accountTestObj{
		w:               w,
		originBalance:   w.GetBalance(),
		originNonce:     w.GetNonce(),
		txHashes:        make(map[string]types.Address),
		spentGas:        big.NewInt(0),
		expectedBalance: w.GetBalance(),
		expectedNonce:   w.GetNonce(),
	}
}

func addSenderAccountTestObj(k string, v *accountTestObj) {
	gmtx.Lock()
	defer gmtx.Unlock()

	senderAccountTestObjMap[k] = v
}

func TestBulkTransfer(t *testing.T) {

	wg := sync.WaitGroup{}

	var allAccountTestObjArr []*accountTestObj
	senderCnt := 0
	for _, w := range wallets {
		if bytes.Compare(w.Address(), validatorWallet.Address()) == 0 {
			continue
		}

		require.NoError(t, w.SyncAccount())

		acctTestObj := newAccountTestObj(w)
		allAccountTestObjArr = append(allAccountTestObjArr, acctTestObj)

		fmt.Println(w.Address(), w.GetNonce(), w.GetBalance())

		if senderCnt < 90 && w.GetBalance().Cmp(big.NewInt(1000000)) >= 0 {
			addSenderAccountTestObj(w.Address().String(), acctTestObj)
			senderCnt++
		}
	}

	for i := len(allAccountTestObjArr); i < 100; i++ {
		newAcctTestObj := newAccountTestObj(client.NewWallet(TESTPASS))
		require.NoError(t, saveRandWallet(newAcctTestObj.w))
		allAccountTestObjArr = append(allAccountTestObjArr, newAcctTestObj)
	}

	t.Logf("TestBulkTransfer - senders: %v, all: %v\n", len(senderAccountTestObjMap), len(allAccountTestObjArr))

	for _, v := range senderAccountTestObjMap {
		wg.Add(1)
		go bulkTransfer(t, &wg, v, allAccountTestObjArr, 50) // 50 txs
	}

	wg.Wait()

	t.Logf("TestBulkTransfer - Check %v accounts ...\n", len(allAccountTestObjArr))

	for _, acctObj := range allAccountTestObjArr {
		//fmt.Println("\tCheck account", acctObj.w.Address())

		require.NoError(t, acctObj.w.SyncAccount())
		require.Equal(t, acctObj.expectedBalance, acctObj.w.GetBalance(), acctObj.w.Address().String())
		require.Equal(t, acctObj.expectedNonce, acctObj.w.GetNonce(), acctObj.w.Address().String())
	}
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, senderAcctTestObj *accountTestObj, receivers []*accountTestObj, cnt int) {
	w := senderAcctTestObj.w
	require.NoError(t, w.Unlock(TESTPASS))

	//t.Logf("bulkTransfer - account: %v, balance: %v, nonce: %v\n", w.Address(), w.GetBalance(), w.GetNonce())

	subWg := sync.WaitGroup{}

	sub, err := client.NewSubscriber("ws://localhost:26657/websocket")
	defer func() {
		sub.Stop()
	}()

	require.NoError(t, err)
	query := fmt.Sprintf("tm.event='Tx' AND tx.sender='%v'", w.Address())
	//fmt.Println("query", query)
	err = sub.Start(query, func(sub *client.Subscriber, result []byte) {

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
		require.Equal(t, gas, big.NewInt(eventDataTx.TxResult.Result.GasUsed))

		subWg.Done()
	})
	require.NoError(t, err)

	maxAmt := new(big.Int).Div(senderAcctTestObj.originBalance, big.NewInt(int64(cnt)))
	maxAmt = new(big.Int).Sub(maxAmt, gas)

	for i := 0; i < cnt; i++ {
		rn := rand.Intn(len(receivers))
		if bytes.Compare(receivers[rn].w.Address(), w.Address()) == 0 {
			rn = (rn + 1) % len(receivers)
		}

		racctState := receivers[rn]
		raddr := racctState.w.Address()

		randAmt := rbytes.RandBigIntN(maxAmt)
		if randAmt.Sign() == 0 {
			randAmt = big.NewInt(1)
			t.Logf("bulkTransfer - from: %v, to: %v, amount: %v\n", w.Address(), raddr, randAmt)
		}
		needAmt := new(big.Int).Add(randAmt, gas)

		subWg.Add(1) // done in subscriber's callback

		ret, err := w.TransferSync(raddr, gas, randAmt)
		require.NoError(t, err)
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log, w.GetNonce(), ret.Hash)

		// todo: this should be locked by mutex
		// record expected state of account
		senderAcctTestObj.txHashes[ret.Hash.String()] = w.Address()
		senderAcctTestObj.spentGas = new(big.Int).Add(senderAcctTestObj.spentGas, gas)
		senderAcctTestObj.expectedBalance = new(big.Int).Sub(senderAcctTestObj.expectedBalance, needAmt)
		senderAcctTestObj.expectedNonce += 1

		racctState.expectedBalance = new(big.Int).Add(racctState.expectedBalance, randAmt)

		fmt.Printf("Send Tx [txHash: %v, from: %v, to: %v, nonce: %v, amt: %v]\n", ret.Hash, w.Address(), racctState.w.Address(), w.GetNonce()+1, randAmt)

		w.AddNonce()
	}

	subWg.Wait()

	wg.Done()
}

func saveRandWallet(w *client.Wallet) error {
	path := filepath.Join(WALKEYDIR, fmt.Sprintf("wk%X.json", w.Address()))
	return w.Save(libs.NewFileWriter(path))
}
