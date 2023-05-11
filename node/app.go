package node

import (
	"fmt"
	"github.com/holiman/uint256"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/cmd/version"
	"github.com/rigochain/rigo-go/ctrlers/account"
	"github.com/rigochain/rigo-go/ctrlers/gov"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/ctrlers/vm/evm"
	"github.com/rigochain/rigo-go/genesis"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcicli "github.com/tendermint/tendermint/abci/client"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	tmver "github.com/tendermint/tendermint/version"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var _ abcitypes.Application = (*RigoApp)(nil)

type RigoApp struct {
	abcitypes.BaseApplication

	currBlockCtx *types2.BlockContext

	metaDB      *MetaDB
	acctCtrler  *account.AcctCtrler
	stakeCtrler *stake.StakeCtrler
	govCtrler   *gov.GovCtrler
	vmCtrler    *evm.EVMCtrler
	txExecutor  *TrxExecutor

	localClient abcicli.Client

	started int32

	logger log.Logger
	mtx    sync.Mutex
}

func NewRigoApp(config *cfg.Config, logger log.Logger) *RigoApp {
	stateDB, err := openMetaDB("rigo_app", config.DBDir())
	if err != nil {
		panic(err)
	}

	govCtrler, err := gov.NewGovCtrler(config, logger)
	if err != nil {
		panic(err)
	}

	acctCtrler, err := account.NewAcctCtrler(config, logger)
	if err != nil {
		panic(err)
	}

	stakeCtrler, err := stake.NewStakeCtrler(config, govCtrler, logger)
	if err != nil {
		panic(err)
	}

	vmCtrler := evm.NewEVMCtrler(config.DBDir(), acctCtrler, logger)

	txExecutor := NewTrxExecutor(runtime.GOMAXPROCS(0), logger)

	return &RigoApp{
		metaDB:      stateDB,
		acctCtrler:  acctCtrler,
		stakeCtrler: stakeCtrler,
		govCtrler:   govCtrler,
		vmCtrler:    vmCtrler,
		txExecutor:  txExecutor,
		logger:      logger,
	}
}

func (ctrler *RigoApp) Start() error {
	if atomic.CompareAndSwapInt32(&ctrler.started, 0, 1) {
		ctrler.txExecutor.Start()
	}
	return nil
}

func (ctrler *RigoApp) Stop() error {
	ctrler.txExecutor.Stop()
	if err := ctrler.acctCtrler.Close(); err != nil {
		return err
	}
	if err := ctrler.stakeCtrler.Close(); err != nil {
		return err
	}
	if err := ctrler.govCtrler.Close(); err != nil {
		return err
	}
	if err := ctrler.metaDB.Close(); err != nil {
		return err
	}
	return nil
}

func (ctrler *RigoApp) SetLocalClient(client abcicli.Client) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// todo: Find out how to solve the following problem.
	// Problem: The 'web3' MUST BE a web3 of CONSENSUS.
	// However, there is no way to know if the 'web3' is for CONSENSUS or not.
	ctrler.localClient = client
}

func (ctrler *RigoApp) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	ctrler.logger.Info("Info", "version", tmver.ABCIVersion, "AppVersion", version.String())

	var appHash bytes.HexBytes
	var lastHeight int64
	ctrler.currBlockCtx = ctrler.metaDB.LastBlockContext()
	if ctrler.currBlockCtx == nil {
		// to ensure backward compatibility
		lastHeight = ctrler.metaDB.LastBlockHeight()
		appHash = ctrler.metaDB.LastBlockAppHash()

		ctrler.currBlockCtx = types2.NewBlockContext(
			abcitypes.RequestBeginBlock{
				Header: tmproto.Header{
					Height: lastHeight,
					Time:   tmtime.Canonical(time.Now()),
				},
			},
			nil, nil, nil)
		ctrler.currBlockCtx.SetAppHash(appHash)
	} else {
		lastHeight = ctrler.currBlockCtx.Height()
		appHash = ctrler.currBlockCtx.AppHash()
	}

	return abcitypes.ResponseInfo{
		Data:             "",
		Version:          tmver.ABCIVersion,
		AppVersion:       version.Uint64(),
		LastBlockHeight:  lastHeight,
		LastBlockAppHash: appHash,
	}
}

// InitChain is called only when the ResponseInfo::LastBlockHeight which is returned in Info() is 0.
func (ctrler *RigoApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	appState := genesis.GenesisAppState{}
	if err := tmjson.Unmarshal(req.AppStateBytes, &appState); err != nil {
		panic(err)
	}

	// todo: check whether 'appHash' is equal to the original hash of the current blockchain network.
	// but how to get the original hash? official web site????
	appHash, err := appState.Hash()
	if err != nil {
		panic(err)
	}

	if xerr := ctrler.govCtrler.InitLedger(&appState); xerr != nil {
		panic(xerr)
	}
	if xerr := ctrler.acctCtrler.InitLedger(&appState); xerr != nil {
		panic(xerr)
	}
	if xerr := ctrler.stakeCtrler.InitLedger(req.Validators); xerr != nil {
		panic(xerr)
	}

	// these values will be saved as state of the consensus engine.
	return abcitypes.ResponseInitChain{
		AppHash: appHash,
	}
}

func (ctrler *RigoApp) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch req.Type {
	case abcitypes.CheckTxType_New:
		txctx, xerr := types2.NewTrxContext(req.Tx,
			0, //ctrler.currBlockCtx.Height()+int64(1),
			0,
			false,
			func(_txctx *types2.TrxContext) xerrors.XError {
				_tx := _txctx.Tx

				_txctx.NeedAmt = new(uint256.Int).Add(_tx.Amount, _tx.Gas)
				_txctx.TrxGovHandler = ctrler.govCtrler
				_txctx.TrxAcctHandler = ctrler.acctCtrler
				_txctx.TrxStakeHandler = ctrler.stakeCtrler
				_txctx.TrxEVMHandler = ctrler.vmCtrler
				_txctx.GovHandler = ctrler.govCtrler
				_txctx.StakeHandler = ctrler.stakeCtrler
				return nil
			})
		if xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			return abcitypes.ResponseCheckTx{
				Code: xerr.Code(),
				Log:  xerr.Error(),
			}
		}

		xerr = ctrler.txExecutor.ExecuteSync(txctx)
		if xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			return abcitypes.ResponseCheckTx{
				Code: xerr.Code(),
				Log:  xerr.Error(),
			}
		}

		return abcitypes.ResponseCheckTx{
			Code:      abcitypes.CodeTypeOK,
			Log:       "",
			Data:      txctx.RetData,
			GasWanted: int64(txctx.Tx.Gas.Uint64()),
			GasUsed:   int64(txctx.GasUsed.Uint64()),
		}
	case abcitypes.CheckTxType_Recheck:
		// do nothing
	}
	return abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK}
}

func (ctrler *RigoApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	if req.Header.Height != ctrler.currBlockCtx.Height()+1 {
		panic(fmt.Errorf("error block height: expected(%v), actural(%v)", ctrler.currBlockCtx.Height()+1, req.Header.Height))
	}

	// save the block fee info. - it will be used for rewarding
	ctrler.currBlockCtx = types2.NewBlockContext(req, ctrler.govCtrler, ctrler.acctCtrler, ctrler.stakeCtrler)

	// todo: implement processing for the evidences (req.ByzantineValidators)

	return abcitypes.ResponseBeginBlock{}
}

func (ctrler *RigoApp) deliverTxSync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {

	txctx, xerr := types2.NewTrxContext(req.Tx,
		ctrler.currBlockCtx.Height(),
		ctrler.currBlockCtx.TimeNano(),
		true,
		func(_txctx *types2.TrxContext) xerrors.XError {
			_tx := _txctx.Tx

			_txctx.TxIdx = ctrler.currBlockCtx.TxsCnt()
			ctrler.currBlockCtx.AddTxsCnt(1)

			_txctx.NeedAmt = new(uint256.Int).Add(_tx.Amount, _tx.Gas)
			_txctx.TrxGovHandler = ctrler.govCtrler
			_txctx.TrxAcctHandler = ctrler.acctCtrler
			_txctx.TrxStakeHandler = ctrler.stakeCtrler
			_txctx.TrxEVMHandler = ctrler.vmCtrler
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			return nil
		})
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		return abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
	}
	xerr = ctrler.txExecutor.ExecuteSync(txctx)
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		return abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
	} else {

		ctrler.currBlockCtx.AddGas(txctx.GasUsed)
		return abcitypes.ResponseDeliverTx{
			Code:      abcitypes.CodeTypeOK,
			GasWanted: int64(txctx.Tx.Gas.Uint64()),
			GasUsed:   int64(txctx.GasUsed.Uint64()),
			Data:      txctx.RetData,
			Events: []abcitypes.Event{
				{
					Type: "tx",
					Attributes: []abcitypes.EventAttribute{
						{Key: []byte(types2.EVENT_ATTR_TXTYPE), Value: []byte(txctx.Tx.TypeString()), Index: true},
						{Key: []byte(types2.EVENT_ATTR_TXSENDER), Value: []byte(txctx.Tx.From.String()), Index: true},
						{Key: []byte(types2.EVENT_ATTR_TXRECVER), Value: []byte(txctx.Tx.To.String()), Index: true},
						{Key: []byte(types2.EVENT_ATTR_ADDRPAIR), Value: []byte(txctx.Tx.From.String() + txctx.Tx.To.String()), Index: true},
					},
				},
			},
		}
	}
}

func (ctrler *RigoApp) deliverTxAsync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	txIdx := ctrler.currBlockCtx.TxsCnt()
	ctrler.currBlockCtx.AddTxsCnt(1)

	txctx, xerr := types2.NewTrxContext(req.Tx,
		ctrler.currBlockCtx.Height(),
		ctrler.currBlockCtx.TimeNano(),
		true,
		func(_txctx *types2.TrxContext) xerrors.XError {
			_tx := _txctx.Tx

			_txctx.TxIdx = txIdx

			_txctx.NeedAmt = new(uint256.Int).Add(_tx.Amount, _tx.Gas)
			_txctx.TrxGovHandler = ctrler.govCtrler
			_txctx.TrxAcctHandler = ctrler.acctCtrler
			_txctx.TrxStakeHandler = ctrler.stakeCtrler
			_txctx.TrxEVMHandler = ctrler.vmCtrler
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			// when the 'tx' is finished, it's called
			_txctx.Callback = func(ctx *types2.TrxContext, xerr xerrors.XError) {
				// it is called from executionRoutine goroutine
				// when execution is finished or error is generated
				response := abcitypes.ResponseDeliverTx{}
				if xerr != nil {
					xerr = xerrors.ErrDeliverTx.Wrap(xerr)
					response.Code = xerr.Code()
					response.Log = xerr.Error()

				} else {
					response.GasWanted = int64(ctx.Tx.Gas.Uint64())
					response.GasUsed = int64(ctx.GasUsed.Uint64())
					response.Data = ctx.RetData
					response.Events = []abcitypes.Event{
						{
							Type: "tx",
							Attributes: []abcitypes.EventAttribute{
								{Key: []byte(types2.EVENT_ATTR_TXTYPE), Value: []byte(ctx.Tx.TypeString()), Index: true},
								{Key: []byte(types2.EVENT_ATTR_TXSENDER), Value: []byte(ctx.Tx.From.String()), Index: true},
								{Key: []byte(types2.EVENT_ATTR_TXRECVER), Value: []byte(ctx.Tx.To.String()), Index: true},
								{Key: []byte(types2.EVENT_ATTR_ADDRPAIR), Value: []byte(ctx.Tx.From.String() + ctx.Tx.To.String()), Index: true},
							},
						},
					}

					ctrler.currBlockCtx.AddGas(ctx.GasUsed)
				}
				ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, ctx.TxIdx, &req, &response)
			}

			return nil
		})
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response := abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
		ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, txIdx, &req, &response)
	}

	xerr = ctrler.txExecutor.ExecuteAsync(txctx)
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response := abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
		ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, txIdx, &req, &response)
	}

	// this return value has no meaning
	return abcitypes.ResponseDeliverTx{}
}

func (ctrler *RigoApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	return ctrler.deliverTxAsync(req)
}

func (ctrler *RigoApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	_ = ctrler.govCtrler.ExecuteBlock(ctrler.currBlockCtx)
	_ = ctrler.acctCtrler.ExecuteBlock(ctrler.currBlockCtx)
	_ = ctrler.stakeCtrler.ExecuteBlock(ctrler.currBlockCtx)

	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: ctrler.currBlockCtx.ValUpdates,
	}
}

func (ctrler *RigoApp) Commit() abcitypes.ResponseCommit {

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	appHash0, ver0, err := ctrler.govCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("RigoApp-Commit", "height", ver0, "appHash0", bytes.HexBytes(appHash0))

	appHash1, ver1, err := ctrler.acctCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("RigoApp-Commit", "height", ver1, "appHash1", bytes.HexBytes(appHash1))

	appHash2, ver2, err := ctrler.stakeCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("RigoApp-Commit", "height", ver2, "appHash2", bytes.HexBytes(appHash2))

	appHash3, ver3, err := ctrler.vmCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("RigoApp-Commit", "height", ver3, "appHash3", bytes.HexBytes(appHash3))

	if ver0 != ver1 || ver1 != ver2 || ver2 != ver3 {
		panic(fmt.Sprintf("Not same versions: gov: %v, account:%v, stake:%v", ver0, ver1, ver2))
	}

	appHash := crypto.DefaultHash(appHash0, appHash1, appHash2, appHash3)
	ctrler.logger.Debug("RigoApp-Commit", "height", ver0, "final hash", bytes.HexBytes(appHash))

	ctrler.currBlockCtx.SetAppHash(appHash)
	ctrler.metaDB.PutLastBlockContext(ctrler.currBlockCtx)
	ctrler.metaDB.PutLastBlockHeight(ver0)
	//ctrler.currBlockCtx = nil

	return abcitypes.ResponseCommit{
		Data: appHash[:],
	}
}
