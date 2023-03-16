package node

import (
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/cmd/version"
	"github.com/rigochain/rigo-go/ctrlers/account"
	"github.com/rigochain/rigo-go/ctrlers/gov"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/genesis"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcicli "github.com/tendermint/tendermint/abci/client"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmver "github.com/tendermint/tendermint/version"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

var _ abcitypes.Application = (*RigoApp)(nil)

type RigoApp struct {
	abcitypes.BaseApplication

	currBlockCtx *types2.BlockContext

	stateDB     *StateDB
	acctCtrler  *account.AcctCtrler
	stakeCtrler *stake.StakeCtrler
	govCtrler   *gov.GovCtrler
	txExecutor  *TrxExecutor

	localClient abcicli.Client

	started int32

	logger log.Logger
	mtx    sync.Mutex
}

func NewRigoApp(config *cfg.Config, logger log.Logger) *RigoApp {
	stateDB, err := openStateDB("rigo_app", config.DBDir())
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

	txExecutor := NewTrxExecutor(runtime.GOMAXPROCS(0), logger)

	return &RigoApp{
		stateDB:     stateDB,
		acctCtrler:  acctCtrler,
		stakeCtrler: stakeCtrler,
		govCtrler:   govCtrler,
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
	if err := ctrler.stateDB.Close(); err != nil {
		return err
	}
	return nil
}

func (ctrler *RigoApp) SetLocalClient(client abcicli.Client) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// todo: Find out how to solve the following problem.
	// Problem: The 'client' MUST BE a client of CONSENSUS.
	// However, there is no way to know if the 'client' is for CONSENSUS or not.
	ctrler.localClient = client
}

func (ctrler *RigoApp) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	ctrler.logger.Info("Info", "version", tmver.ABCIVersion, "AppVersion", version.String())

	return abcitypes.ResponseInfo{
		Data:             "",
		Version:          tmver.ABCIVersion,
		AppVersion:       version.Uint64(),
		LastBlockHeight:  ctrler.stateDB.LastBlockHeight(),
		LastBlockAppHash: ctrler.stateDB.LastBlockAppHash(),
	}
}

// InitChain is called only when the ResponseInfo::LastBlockHeight which is returned in Info() is 0.
func (ctrler *RigoApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	appState := genesis.GenesisAppState{}
	if err := json.Unmarshal(req.AppStateBytes, &appState); err != nil {
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
	response := abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK}

	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	switch req.Type {
	case abcitypes.CheckTxType_New:
		if txctx, xerr := types2.NewTrxContext(req.Tx,
			ctrler.stateDB.LastBlockHeight()+int64(1),
			false,
			func(_txctx *types2.TrxContext) xerrors.XError {
				_tx := _txctx.Tx

				_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
				_txctx.GovHandler = ctrler.govCtrler
				_txctx.AcctHandler = ctrler.acctCtrler
				_txctx.StakeHandler = ctrler.stakeCtrler
				_txctx.GovHelper = ctrler.govCtrler
				_txctx.StakeHelper = ctrler.stakeCtrler
				return nil
			}); xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			response.Code = xerr.Code()
			response.Log = xerr.Error()
		} else if xerr := ctrler.txExecutor.ExecuteSync(txctx); xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			response.Code = xerr.Code()
			response.Log = xerr.Error()
		} else {
			response.GasWanted = txctx.Tx.Gas.Int64()
			response.GasUsed = txctx.GasUsed.Int64()
		}
	case abcitypes.CheckTxType_Recheck:
		// do nothing
	}
	return response
}

func (ctrler *RigoApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	if req.Header.Height != ctrler.stateDB.LastBlockHeight()+1 {
		panic(fmt.Errorf("error block height: expected(%v), actural(%v)", ctrler.stateDB.LastBlockHeight()+1, req.Header.Height))
	}

	// save the block fee info. - it will be used for rewarding
	ctrler.currBlockCtx = &types2.BlockContext{
		BlockInfo:   req,
		TxsCnt:      0,
		Fee:         big.NewInt(0),
		GovHelper:   ctrler.govCtrler,
		AcctHelper:  ctrler.acctCtrler,
		StakeHelper: ctrler.stakeCtrler,
	}

	// todo: implement processing for the evidences (req.ByzantineValidators)

	// todo: change new callback function of ctrler.localClient
	// the new callback should wait that a DeliverTx is completed.

	return abcitypes.ResponseBeginBlock{}
}

func (ctrler *RigoApp) deliverTxSync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	response := abcitypes.ResponseDeliverTx{}

	if txctx, xerr := types2.NewTrxContext(req.Tx,
		ctrler.stateDB.LastBlockHeight()+int64(1),
		true,
		func(_txctx *types2.TrxContext) xerrors.XError {
			_tx := _txctx.Tx

			_txctx.TxIdx = ctrler.currBlockCtx.TxsCnt
			ctrler.currBlockCtx.TxsCnt++

			_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.AcctHandler = ctrler.acctCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			_txctx.GovHelper = ctrler.govCtrler
			_txctx.StakeHelper = ctrler.stakeCtrler
			return nil
		}); xerr != nil {

		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response.Code = xerr.Code()
		response.Log = xerr.Error()
	} else if xerr := ctrler.txExecutor.ExecuteSync(txctx); xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response.Code = xerr.Code()
		response.Log = xerr.Error()
	} else {

		_ = ctrler.currBlockCtx.Fee.Add(ctrler.currBlockCtx.Fee, txctx.GasUsed)
		response.GasWanted = txctx.Tx.Gas.Int64()
		response.GasUsed = txctx.GasUsed.Int64()

		response.Events = []abcitypes.Event{
			{
				Type: "tx",
				Attributes: []abcitypes.EventAttribute{
					{Key: []byte(types2.EVENT_ATTR_TXTYPE), Value: []byte(txctx.Tx.TypeString()), Index: true},
					{Key: []byte(types2.EVENT_ATTR_TXSENDER), Value: []byte(txctx.Tx.From.String()), Index: true},
					{Key: []byte(types2.EVENT_ATTR_TXRECVER), Value: []byte(txctx.Tx.To.String()), Index: true},
					{Key: []byte(types2.EVENT_ATTR_ADDRPAIR), Value: []byte(txctx.Tx.From.String() + txctx.Tx.To.String()), Index: true},
				},
			},
		}
	}
	return response
}

func (ctrler *RigoApp) deliverTxAsync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	txIdx := ctrler.currBlockCtx.TxsCnt
	ctrler.currBlockCtx.TxsCnt++

	response := abcitypes.ResponseDeliverTx{}

	if txctx, xerr := types2.NewTrxContext(req.Tx,
		ctrler.stateDB.LastBlockHeight()+int64(1),
		true,
		func(_txctx *types2.TrxContext) xerrors.XError {
			_tx := _txctx.Tx

			_txctx.TxIdx = txIdx

			_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.AcctHandler = ctrler.acctCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			_txctx.GovHelper = ctrler.govCtrler
			_txctx.StakeHelper = ctrler.stakeCtrler
			// when the 'tx' is finished, it's called
			_txctx.Callback = func(ctx *types2.TrxContext, xerr xerrors.XError) {
				// todo: notify that the tx is finisehd
				// it is called in the executionRoutine goroutine
				if xerr != nil {
					xerr = xerrors.ErrDeliverTx.Wrap(xerr)
					response.Code = xerr.Code()
					response.Log = xerr.Error()

				} else {
					ctrler.mtx.Lock()
					defer ctrler.mtx.Unlock()

					_ = ctrler.currBlockCtx.Fee.Add(ctrler.currBlockCtx.Fee, ctx.GasUsed)

					response.GasWanted = ctx.Tx.Gas.Int64()
					response.GasUsed = ctx.GasUsed.Int64()
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
				}
				ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, ctx.TxIdx, &req, &response)
			}
			return nil
		}); xerr != nil {

		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response.Code = xerr.Code()
		response.Log = xerr.Error()

		ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, txIdx, &req, &response)
	} else if xerr := ctrler.txExecutor.ExecuteAsync(txctx); xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		response.Code = xerr.Code()
		response.Log = xerr.Error()

		ctrler.localClient.(*rigoLocalClient).OnTrxExecFinished(ctrler.localClient, txIdx, &req, &response)
	}
	return response
}

func (ctrler *RigoApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	// todo: call ctrler.deliverTxAsync() when is implemented.
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
	ctrler.logger.Debug("RigoApp-Commit", "height", ver0, "appHash1", bytes.HexBytes(appHash1))

	appHash2, ver2, err := ctrler.stakeCtrler.Commit()
	if err != nil {
		panic(err)
	}
	ctrler.logger.Debug("RigoApp-Commit", "height", ver0, "appHash2", bytes.HexBytes(appHash2))

	if ver0 != ver1 || ver1 != ver2 {
		panic(fmt.Sprintf("Not same versions: gov: %v, account:%v, stake:%v", ver0, ver1, ver2))
	}

	appHash := crypto.DefaultHash(appHash0, appHash1, appHash2)
	ctrler.logger.Debug("RigoApp-Commit", "height", ver0, "final hash", bytes.HexBytes(appHash))

	ctrler.stateDB.PutLastBlockHeight(ver0)
	ctrler.stateDB.PutLastBlockAppHash(appHash[:])
	ctrler.stateDB.PutLastBlockContext(ctrler.currBlockCtx)
	ctrler.currBlockCtx = nil

	return abcitypes.ResponseCommit{
		Data: appHash[:],
	}
}
