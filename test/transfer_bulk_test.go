package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/rigochain/rigo-go/libs/client"
	rbytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"math/big"
	"sync"
	"testing"
)

type accountState struct {
	w *client.Wallet

	originBalance *big.Int
	originNonce   uint64

	spentGas *big.Int
	txHashes []rbytes.HexBytes

	expectedBalance *big.Int
	expectedNonce   uint64
}

func newAccountState(w *client.Wallet) *accountState {
	return &accountState{
		w:               w,
		originBalance:   w.GetBalance(),
		originNonce:     w.GetNonce(),
		spentGas:        big.NewInt(0),
		expectedBalance: w.GetBalance(),
		expectedNonce:   w.GetNonce(),
	}
}

var accountStates map[string]*accountState
var gmtx sync.Mutex

func addAccountState(k string, v *accountState) {
	gmtx.Lock()
	defer gmtx.Unlock()

	accountStates[k] = v
}

func TestBulkTransfer(t *testing.T) {
	accountStates = make(map[string]*accountState, len(wallets))

	wg := sync.WaitGroup{}

	for i, w := range wallets {

		if bytes.Compare(w.Address(), validatorWallet.Address()) == 0 {
			continue
		}

		require.NoError(t, w.SyncAccount())

		fmt.Println("index", i, "address", w.Address(), w.GetNonce(), w.GetBalance())

		acctState := newAccountState(w)
		accountStates[w.Address().String()] = acctState

		wg.Add(1)
		go bulkTransfer(t, &wg, acctState, 10)
	}

	wg.Wait()

	fmt.Printf("Check accounts[%v] state...\n", len(accountStates))

	//for _, acctRet := range accountStates {
	//	for _, txHash := range acctRet.txHashes {
	//		txRet, err := waitTrxResult(txHash, 20)
	//		require.NoError(t, err)
	//
	//		// check tx result
	//		require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log, txRet.TxDetail.Nonce, txHash)
	//		require.EqualValues(t, txHash, txRet.Hash)
	//		require.Equal(t, gas, txRet.TxDetail.Gas)
	//		//require.Equal(t, randAmt, txRet.TxDetail.Amount)
	//	}
	//}

	for _, acctRet := range accountStates {
		fmt.Println("\tCheck account", acctRet.w.Address())

		require.NoError(t, acctRet.w.SyncAccount())
		require.Equal(t, acctRet.expectedBalance, acctRet.w.GetBalance(), acctRet.w.Address().String())
		require.Equal(t, acctRet.expectedNonce, acctRet.w.GetNonce(), acctRet.w.Address().String())

		require.NotEqual(t, acctRet.expectedBalance, acctRet.originBalance, acctRet.w.Address().String())
		if acctRet.originBalance.Cmp(big.NewInt(0)) != 0 {
			// receiver account
			require.NotEqual(t, acctRet.expectedNonce, acctRet.originNonce, acctRet.w.Address().String())
		}
	}
}

func bulkTransfer(t *testing.T, wg *sync.WaitGroup, acctState *accountState, cnt int) {
	w := acctState.w
	require.NoError(t, w.Unlock(TESTPASS))

	maxAmt := new(big.Int).Div(acctState.originBalance, big.NewInt(int64(cnt)))
	maxAmt = new(big.Int).Sub(maxAmt, gas)

	//fmt.Println("\tstart account:", w.Address(), w.GetBalance(), w.GetNonce())

	subWg := sync.WaitGroup{}

	sub, err := client.NewSubscriber("ws://localhost:26657/websocket")
	defer sub.Stop()

	require.NoError(t, err)
	query := fmt.Sprintf("tm.event='Tx' AND tx.sender='%v'", w.Address())
	fmt.Println("query", query)
	err = sub.Watch(query, func(sub *client.Subscriber, result []byte) {

		event := &coretypes.ResultEvent{}
		err := tmjson.Unmarshal(result, event)
		require.NoError(t, err)

		found := false
		txHash, err := hex.DecodeString(event.Events["tx.hash"][0])
		require.NoError(t, err)

		for _, h := range acctState.txHashes {
			if bytes.Compare(txHash, h) == 0 {
				found = true
				fmt.Println("Found Tx", event.Events["tx.hash"])
			}
		}
		subWg.Done()
		require.True(t, found)

		eventDataTx := event.Data.(types.EventDataTx)
		require.Equal(t, xerrors.ErrCodeSuccess, eventDataTx.TxResult.Result.Code)
		require.Equal(t, gas, big.NewInt(eventDataTx.TxResult.Result.GasUsed))

	})
	require.NoError(t, err)

	for i := 0; i < cnt; i++ {

		racctState := newAccountState(client.NewWallet(TESTPASS))
		raddr := racctState.w.Address()

		addAccountState(raddr.String(), racctState)

		randAmt := rbytes.RandBigIntN(maxAmt)
		needAmt := new(big.Int).Add(randAmt, gas)

		subWg.Add(1) // done in subscriber's callback

		ret, err := w.TransferSync(raddr, gas, randAmt)
		require.NoError(t, err)
		require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log, w.GetNonce(), ret.Hash)

		////wait tx
		//txRet, err := waitTrxResult(txHash, 20)
		//require.NoError(t, err)
		//
		//// check tx result
		//require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code)
		//require.Equal(t, txHash, txRet.Hash)
		//require.Equal(t, gas, txRet.TxDetail.Gas)
		//require.Equal(t, randAmt, txRet.TxDetail.Amount)

		// record expected state of account
		acctState.txHashes = append(acctState.txHashes, rbytes.HexBytes(ret.Hash))
		acctState.spentGas = new(big.Int).Add(acctState.spentGas, gas)
		acctState.expectedBalance = new(big.Int).Sub(acctState.expectedBalance, needAmt)
		acctState.expectedNonce += 1

		racctState.expectedBalance = new(big.Int).Add(racctState.expectedBalance, randAmt)

		fmt.Printf("\tTx [txHash: %v, from: %v, to: %v, nonce: %v, amt: %v]\n", ret.Hash, w.Address(), racctState.w.Address(), w.GetNonce()+1, randAmt)

		w.AddNonce()
	}

	subWg.Wait()

	wg.Done()

	//fmt.Println("\tgoodbye", w.Address())

}
