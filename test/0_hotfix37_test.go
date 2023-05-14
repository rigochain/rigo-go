package test

import (
	"bytes"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStaking2GenesisValidator(t *testing.T) {
	rweb3 := randRigoWeb3()

	valWal := validatorWallets[0]
	require.NoError(t, valWal.SyncAccount(rweb3))
	require.NoError(t, valWal.Unlock(defaultRpcNode.Pass))

	ret, err := valWal.StakingSync(valWal.Address(), gas, uint256.NewInt(1000000000000000000), rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	txRet, err := waitTrxResult(ret.Hash, 15, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, txRet.TxResult.Code, txRet.TxResult.Log)

	stakes, err := rweb3.GetStakes(valWal.Address())
	require.NoError(t, err)
	require.Equal(t, 2, len(stakes), stakes)

	found := false
	for _, s := range stakes {
		if bytes.Compare(ret.Hash, s.TxHash) == 0 {
			found = true
			break
		}
	}
	require.True(t, found)
}
