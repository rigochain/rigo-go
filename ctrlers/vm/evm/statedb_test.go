package evm

//func TestStateDBWrapper_Prepare(t *testing.T) {
//	deployInput, err := abiERC20Contract.Pack("", "TokenOnRigo", "TOR")
//	require.NoError(t, err)
//
//	// deploy code = contract byte code + input parameters
//	deployInput = append(buildInfo.Bytecode, deployInput...)
//
//	// make transaction
//	fromAcct := acctHandler.walletsArr[0].GetAccount()
//	to := types.ZeroAddress()
//
//	bctx := ctrlertypes.NewBlockContext(abcitypes.RequestBeginBlock{Header: tmproto.Header{Height: rigoEVM.lastBlockHeight + 1}}, nil, &acctHandler, nil)
//	_, xerr := rigoEVM.BeginBlock(bctx)
//	require.NoError(t, xerr)
//
//	txctx := &ctrlertypes.TrxContext{
//		Height:     bctx.Height(),
//		BlockTime:  time.Now().Unix(),
//		TxHash:     bytes2.RandBytes(32),
//		Tx:         web3.NewTrxContract(fromAcct.Address, to, fromAcct.GetNonce(), 3_000_000, uint256.NewInt(10_000_000_000), uint256.NewInt(0), deployInput),
//		TxIdx:      1,
//		Exec:       true,
//		Sender:     fromAcct,
//		Receiver:   nil,
//		GasUsed:    0,
//		GovHandler: govParams,
//	}
//
//	xerr = rigoEVM.ExecuteTrx(txctx)
//	require.NoError(t, xerr)
//
//	contractAddr = txctx.RetData
//	fmt.Println("Deployed", "contract address", contractAddr)
//
//	//
//	// update balance of RIGO Ledger
//	fromAcct.SetBalance(uint256.NewInt(0))
//	require.NoError(t, xerr)
//
//	xerr = rigoEVM.ExecuteTrx(txctx)
//	require.NoError(t, xerr)
//
//	_, height, xerr := rigoEVM.Commit()
//	require.NoError(t, xerr)
//	fmt.Println("TestDeploy", "Commit block", height)
//}
