package account_test

import (
	"encoding/json"
	"fmt"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	addr0 = libs.RandAddress()
	dbDir = filepath.Join(os.TempDir(), "account-ctrler-test")
)

func TestAccountMapMarshal(t *testing.T) {
	addrs := make([]types.Address, 10)
	for i, _ := range addrs {
		addrs[i] = libs.RandAddress()
	}

	accts := &struct {
		Map1 map[types.AcctKey]*account.Account `json:"map1"`
	}{
		Map1: make(map[types.AcctKey]*account.Account),
	}
	for _, addr := range addrs {
		accts.Map1[types.ToAcctKey(addr)] = &account.Account{}
	}

	_, err := json.Marshal(accts)
	require.NoError(t, err)

	// how to check that the key is ordered ???
}

func TestAccountCtrler_Find0(t *testing.T) {
	os.RemoveAll(dbDir)

	ctrler, err := account.NewAccountCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)

	// when exec == false
	acct := ctrler.FindAccount(addr0, false)
	require.Nil(t, acct)

	acct = ctrler.FindOrNewAccount(addr0, false)
	require.NotNil(t, acct)
	require.Equal(t, addr0, acct.GetAddress())
	require.Equal(t, uint64(0), acct.GetNonce())

	// when exec == true
	acct = ctrler.FindAccount(addr0, true)
	require.Nil(t, acct)

	acct = ctrler.FindOrNewAccount(addr0, true)
	require.NotNil(t, acct)

	acct.AddNonce()
	acct.AddBalance(big.NewInt(10))

	// save execAccounts
	_, _, err = ctrler.Commit()
	require.NoError(t, err)

	err = ctrler.Close()
	require.NoError(t, err)

	ctrler2, err := account.NewAccountCtrler(dbDir, log.NewNopLogger())
	require.NoError(t, err)

	acct2 := ctrler2.FindAccount(addr0, false)
	require.NotNil(t, acct2)

	require.Equal(t, addr0, acct2.GetAddress())
	require.Equal(t, uint64(1), acct2.GetNonce())
	require.Equal(t, big.NewInt(10), acct2.GetBalance())

	require.NoError(t, ctrler2.Close())

	os.RemoveAll(dbDir)
}

func TestAccountCtrler_AppHash(t *testing.T) {
	var addrList [100]types.Address
	for i := 0; i < len(addrList); i++ {
		addrList[i] = libs.RandAddress()
	}

	wg := sync.WaitGroup{}
	nodeCnt := 20
	appHashes := make([][]byte, nodeCnt)
	blockNums := make([]int64, nodeCnt)

	for i := 0; i < nodeCnt; i++ {
		wg.Add(1)
		go func(idx int) {
			_dbDir := filepath.Join(os.TempDir(), fmt.Sprintf("account-apphash-test_%d", idx))
			os.RemoveAll(_dbDir)

			ctrler, err := account.NewAccountCtrler(_dbDir, log.NewNopLogger())
			require.NoError(t, err, "dbDir", _dbDir)

			for j := 0; j < len(addrList); j++ {
				acct := ctrler.FindOrNewAccount(addrList[j], true)
				require.NotNil(t, acct)
				acct.AddNonce()
				acct.AddBalance(big.NewInt(int64(j * 1000000)))
			}

			appHash, ver, err := ctrler.Commit()
			require.NoError(t, err)

			appHashes[idx] = appHash
			blockNums[idx] = ver

			os.RemoveAll(_dbDir)
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := 1; i < len(appHashes); i++ {
		require.Equal(t, appHashes[0], appHashes[i])
		require.Equal(t, blockNums[0], blockNums[i])
	}

	os.RemoveAll(dbDir)
}
