package test

// Disable test case
// when staking is success, the blockchain node will stop.

//func TestStaking2GenesisValidator(t *testing.T) {
//	rweb3 := randRigoWeb3()
//	govRule, err := rweb3.GetRule()
//	require.NoError(t, err)
//
//	valWal := validatorWallets[0]
//	require.NoError(t, valWal.SyncAccount(rweb3))
//	require.NoError(t, valWal.Unlock(defaultRpcNode.Pass))
//
//	valStakes0, err := rweb3.GetDelegatee(valWal.Address())
//	require.NoError(t, err)
//
//	//fmt.Println("valStake0.SelfAmount", valStakes0.SelfAmount.Dec())
//
//	amtStake := uint256.NewInt(1000000000000000000)
//	ret, err := valWal.StakingCommit(valWal.Address(), defGas, defGasPrice, amtStake, rweb3)
//	require.NoError(t, err)
//	require.NotEqual(t, xerrors.ErrCodeSuccess, ret.CheckTx.Code)
//	require.Contains(t, ret.CheckTx.Log, "too small stake to become validator")
//
//	amtStake = new(uint256.Int).Sub(govRule.MinValidatorStake(), types.PowerToAmount(valStakes0.SelfPower))
//	ret, err = valWal.StakingCommit(valWal.Address(), defGas, defGasPrice, amtStake, rweb3)
//	require.NoError(t, err)
//	require.Equal(t, xerrors.ErrCodeSuccess, ret.DeliverTx.Code, ret.DeliverTx.Log)
//
//	stakes, err := rweb3.GetStakes(valWal.Address())
//	require.NoError(t, err)
//	require.Equal(t, 2, len(stakes), stakes)
//
//	found := false
//	for _, s := range stakes {
//		if bytes.Compare(ret.Hash, s.TxHash) == 0 {
//			found = true
//			break
//		}
//	}
//	require.True(t, found)
//
//	valStakes1, err := rweb3.GetDelegatee(valWal.Address())
//	require.NoError(t, err)
//	require.Equal(t,
//		new(uint256.Int).Add(types.PowerToAmount(valStakes0.TotalPower), amtStake),
//		types.PowerToAmount(valStakes1.TotalPower))
//	require.Equal(t, valStakes1.TotalPower,
//		valStakes1.SumPower())
//
//	//fmt.Println("valStakes1.SelfAmount", valStakes1.SelfAmount.Dec())
//
//}
