package account

import (
	"errors"
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

var (
	acctCtrlers []*AcctCtrler
	addrs       []types.Address
)

func init() {
	if err := initialize(); err != nil {
		panic(err)
	}
}

func initialize() error {
	acctCtrlers = nil
	for i := 0; i < 10; i++ {
		config := cfg.DefaultConfig()
		config.DBPath = filepath.Join(os.TempDir(), fmt.Sprintf("testnode-%d", i))
		_ = os.RemoveAll(config.DBDir())

		ctrler, err := NewAcctCtrler(config, tmlog.NewNopLogger())
		if err != nil {
			return err
		}
		acctCtrlers = append(acctCtrlers, ctrler)
	}

	addrs = nil
	for i := 0; i < 100; i++ {
		addr := bytes.RandBytes(types.AddrSize)
		for j := 0; j < len(acctCtrlers); j++ {
			acct := acctCtrlers[j].FindOrNewAccount(addr, true)
			acct.AddBalance(big.NewInt(10000000))
			acctCtrlers[j].setAccountCommittable(acct, true)
		}
		addrs = append(addrs, addr)
	}
	if _, _, err := commit(); err != nil {
		return err
	}

	return nil
}

func commit() ([]byte, int64, error) {
	var preAppHash, appHash []byte
	var preVer, ver int64
	var xerr xerrors.XError
	for j := 0; j < len(acctCtrlers); j++ {
		appHash, ver, xerr = acctCtrlers[j].Commit()
		if xerr != nil {
			return nil, -1, xerr
		}
		if preAppHash != nil && bytes.Compare(preAppHash, appHash) != 0 {
			return nil, -1, errors.New("appHash is not same")
		}
		if preVer != 0 && preVer != ver {
			return nil, -1, errors.New("version is not same")
		}
		preAppHash = appHash
		preVer = ver
	}
	return appHash, ver, nil
}

func simuRand(n int) error {
	for i := 0; i < n; i++ {
		r0, r1, r2 := rand.Intn(len(addrs)), rand.Intn(len(addrs)), rand.Intn(len(acctCtrlers))
		addr0, addr1, ctrler := addrs[r0], addrs[r1], acctCtrlers[r2]

		if err := ctrler.Transfer(addr0, addr1, big.NewInt(1), false); err != nil {
			return err
		}
	}
	return nil
}
func execRand(n int) error {
	for i := 0; i < n; i++ {
		r0, r1, r2 := rand.Intn(len(addrs)), rand.Intn(len(addrs)), rand.Intn(len(acctCtrlers))
		addr0, addr1, ctrler := addrs[r0], addrs[r1], acctCtrlers[r2]

		if err := ctrler.Transfer(addr0, addr1, big.NewInt(1), true); err != nil {
			return err
		}
	}
	return nil
}

func execSame(n int) error {
	for i := 0; i < n; i++ {
		r0, r1 := rand.Intn(len(addrs)), rand.Intn(len(addrs))
		addr0, addr1 := addrs[r0], addrs[r1]

		for j := 0; j < len(acctCtrlers); j++ {

			if err := acctCtrlers[j].Transfer(addr0, addr1, big.NewInt(1), true); err != nil {
				return err
			}
		}
	}
	return nil
}

func readRand(n int) error {
	for i := 0; i < n; i++ {
		r0, r1 := rand.Intn(len(addrs)), rand.Intn(len(acctCtrlers))
		addr0, ctrler := addrs[r0], acctCtrlers[r1]

		// it makes ledger tree's cash to be dirty
		if acct := ctrler.ReadAccount(addr0); acct == nil {
			return xerrors.ErrNotFoundAccount
		}
	}
	return nil
}

func TestAcctCtrler_Commit(t *testing.T) {
	var preHash []byte
	var preVer int64

	for i := 0; i < 100; i++ {
		require.NoError(t, simuRand(100))
		require.NoError(t, readRand(100))
		h, v, e := commit()
		require.NoError(t, e)

		if preHash != nil {
			require.Equal(t, preHash, h)
		}
		preHash = h

		if preVer != 0 {
			require.Equal(t, preVer+1, v)
		}
		preVer = v
	}

	require.NoError(t, initialize())
	for i := 0; i < 1000; i++ {
		require.NoError(t, simuRand(100))
		require.NoError(t, execRand(100))
		require.NoError(t, readRand(100))

		_, _, e := commit()
		require.Error(t, e)
	}

	require.NoError(t, initialize())
	preHash = nil
	preVer = 0

	for i := 2; i < 1000; i++ {
		require.NoError(t, simuRand(100))
		require.NoError(t, readRand(10))
		require.NoError(t, execSame(100))
		require.NoError(t, readRand(10))

		h, v, e := commit()
		require.NoError(t, e)

		if preHash != nil {
			require.NotEqual(t, preHash, h)
		}
		preHash = h

		if preVer != 0 {
			require.Equal(t, preVer+1, v)
		}
		preVer = v
	}
}
