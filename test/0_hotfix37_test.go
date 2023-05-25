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
	govRule, err := rweb3.GetRule()
	require.NoError(t, err)

	valWal := validatorWallets[0]
	require.NoError(t, valWal.SyncAccount(rweb3))
	require.NoError(t, valWal.Unlock(defaultRpcNode.Pass))

	valStakes0, err := rweb3.GetDelegatee(valWal.Address())
	require.NoError(t, err)

	//fmt.Println("valStake0.SelfAmount", valStakes0.SelfAmount.Dec())

	amtStake := uint256.NewInt(1000000000000000000)
	ret, err := valWal.StakingSync(valWal.Address(), gas10, amtStake, rweb3)
	require.NoError(t, err)
	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.Code)
	require.Contains(t, ret.Log, "too small stake to become validator")

	amtStake = new(uint256.Int).Sub(govRule.MinValidatorStake(), valStakes0.SelfAmount)
	ret, err = valWal.StakingSync(valWal.Address(), gas10, amtStake, rweb3)
	require.NoError(t, err)
	require.Equal(t, xerrors.ErrCodeSuccess, ret.Code, ret.Log)

	txRet, err := waitTrxResult(ret.Hash, 30, rweb3)
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

	valStakes1, err := rweb3.GetDelegatee(valWal.Address())
	require.NoError(t, err)
	require.Equal(t,
		new(uint256.Int).Add(valStakes0.GetTotalAmount(), amtStake),
		valStakes1.GetTotalAmount())
	require.Equal(t, valStakes1.GetTotalAmount(),
		valStakes1.SumAmount())

	//fmt.Println("valStakes1.SelfAmount", valStakes1.SelfAmount.Dec())

}
