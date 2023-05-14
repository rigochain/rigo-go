package test

import (
	"fmt"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestIssue32(t *testing.T) {
	wg := sync.WaitGroup{}

	rweb3 := randRigoWeb3()

	var allAcctHelpers []*acctObj
	senderCnt := 0

	for _, w := range wallets {
		if isValidatorWallet(w) {
			continue
		}

		require.NoError(t, w.SyncAccount(rweb3))

		acctTestObj := newAcctObj(w)
		allAcctHelpers = append(allAcctHelpers, acctTestObj)

		if w.GetBalance().Cmp(uint256.NewInt(1000000)) >= 0 {
			addSenderAcctHelper(w.Address().String(), acctTestObj)
			senderCnt++
		}

		if senderCnt >= 90 {
			break
		}
	}

	randRecvAcct := web3.NewWallet(defaultRpcNode.Pass)
	receiverHelper := newAcctObj(randRecvAcct)
	allAcctHelpers = append(allAcctHelpers, receiverHelper)

	randRecvAcct1 := web3.NewWallet(defaultRpcNode.Pass)
	receiverHelper1 := newAcctObj(randRecvAcct1)
	allAcctHelpers = append(allAcctHelpers, receiverHelper1)

	for _, v := range senderAcctObjs {
		wg.Add(1)
		go bulkTransfer(t, &wg, v, []*acctObj{receiverHelper, receiverHelper1}, 50) // 50 txs
	}

	wg.Wait()

	for _, acctObj := range allAcctHelpers {
		require.NoError(t, acctObj.w.SyncAccount(rweb3))
		require.Equal(t, acctObj.expectedBalance, acctObj.w.GetBalance(), acctObj.w.Address().String())
		require.Equal(t, acctObj.expectedNonce, acctObj.w.GetNonce(), acctObj.w.Address().String())

		fmt.Println("\tCheck account", acctObj.w.Address(), acctObj.expectedNonce, acctObj.expectedBalance, acctObj.w.GetBalance())
	}

	clearSenderAcctHelper()
}
