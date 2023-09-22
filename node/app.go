package node

import (
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/cmd/version"
	"github.com/rigochain/rigo-go/ctrlers/account"
	"github.com/rigochain/rigo-go/ctrlers/gov"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	rctypes "github.com/rigochain/rigo-go/ctrlers/types"
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
	"sync"
	"sync/atomic"
	"time"
)

var _ abcitypes.Application = (*RigoApp)(nil)

type RigoApp struct {
	abcitypes.BaseApplication

	lastBlockCtx *rctypes.BlockContext
	nextBlockCtx *rctypes.BlockContext

	metaDB      *MetaDB
	acctCtrler  *account.AcctCtrler
	stakeCtrler *stake.StakeCtrler
	govCtrler   *gov.GovCtrler
	vmCtrler    *evm.EVMCtrler
	txExecutor  *TrxExecutor

	localClient abcicli.Client
	rootConfig  *cfg.Config

	started int32
	logger  log.Logger
	mtx     sync.Mutex
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

	// the first parameter of NewTrxExecutor `n` is 0,
	// because the parallel tx-processing is not used
	txExecutor := NewTrxExecutor(0 /*runtime.GOMAXPROCS(0)*/, logger)

	return &RigoApp{
		metaDB:      stateDB,
		acctCtrler:  acctCtrler,
		stakeCtrler: stakeCtrler,
		govCtrler:   govCtrler,
		vmCtrler:    vmCtrler,
		txExecutor:  txExecutor,
		rootConfig:  config,
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
	ctrler.lastBlockCtx = ctrler.metaDB.LastBlockContext()
	if ctrler.lastBlockCtx == nil {
		// to ensure backward compatibility
		lastHeight = ctrler.metaDB.LastBlockHeight()
		appHash = ctrler.metaDB.LastBlockAppHash()

		ctrler.lastBlockCtx = rctypes.NewBlockContext(
			abcitypes.RequestBeginBlock{
				Header: tmproto.Header{
					Height: lastHeight,
					Time:   tmtime.Canonical(time.Now()),
				},
			},
			nil, nil, nil)
		ctrler.lastBlockCtx.SetAppHash(appHash)
	} else {
		lastHeight = ctrler.lastBlockCtx.Height()
		appHash = ctrler.lastBlockCtx.AppHash()
	}

	// get chain_id
	ctrler.rootConfig.ChainID = ctrler.metaDB.ChainID()

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
	// set and put chain_id
	if req.GetChainId() == "" {
		panic("there is no chain_id")
	}
	ctrler.rootConfig.ChainID = req.GetChainId()
	_ = ctrler.metaDB.PutChainID(ctrler.rootConfig.ChainID)

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
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	if xerr := ctrler.acctCtrler.InitLedger(&appState); xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}

	// validator - initial stakes
	initStakes := make([]*stake.InitStake, len(req.Validators))
	for i, val := range req.Validators {
		pubBytes := val.PubKey.GetSecp256K1()
		addr, xerr := crypto.PubBytes2Addr(pubBytes)
		if xerr != nil {
			ctrler.logger.Error("RigoApp", "error", xerr)
			panic(xerr)
		}
		s0 := stake.NewStakeWithPower(
			addr, addr, // self staking
			val.Power,
			1,
			bytes.ZeroBytes(32), // 0x00... txhash
		)
		initStakes[i] = &stake.InitStake{
			pubBytes,
			[]*stake.Stake{s0},
		}

		// Generate account of validator,
		// if validator account is not initialized at acctCtrler.InitLedger,

		if ctrler.acctCtrler.FindOrNewAccount(addr, true) == nil {
			panic("fail to create account of validator")
		}
	}

	if xerr := ctrler.stakeCtrler.InitLedger(initStakes); xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
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
		txctx, xerr := rctypes.NewTrxContext(req.Tx,
			ctrler.lastBlockCtx.Height()+int64(1), // issue #39: set block number expected to include current tx.
			ctrler.lastBlockCtx.ExpectedNextBlockTimeSeconds(ctrler.rootConfig.Consensus.CreateEmptyBlocksInterval), // issue #39: set block time expected to be executed.
			false,
			func(_txctx *rctypes.TrxContext) xerrors.XError {
				_txctx.TrxGovHandler = ctrler.govCtrler
				_txctx.TrxAcctHandler = ctrler.acctCtrler
				_txctx.TrxStakeHandler = ctrler.stakeCtrler
				_txctx.TrxEVMHandler = ctrler.vmCtrler
				_txctx.GovHandler = ctrler.govCtrler
				_txctx.AcctHandler = ctrler.acctCtrler
				_txctx.StakeHandler = ctrler.stakeCtrler
				_txctx.ChainID = ctrler.rootConfig.ChainID
				return nil
			})
		if xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			ctrler.logger.Error("CheckTx", "error", xerr)
			return abcitypes.ResponseCheckTx{
				Code: xerr.Code(),
				Log:  xerr.Error(),
			}
		}

		xerr = ctrler.txExecutor.ExecuteSync(txctx)
		if xerr != nil {
			xerr = xerrors.ErrCheckTx.Wrap(xerr)
			ctrler.logger.Error("CheckTx", "error", xerr)
			return abcitypes.ResponseCheckTx{
				Code: xerr.Code(),
				Log:  xerr.Error(),
			}
		}

		return abcitypes.ResponseCheckTx{
			Code:      abcitypes.CodeTypeOK,
			Log:       "",
			Data:      txctx.RetData,
			GasWanted: int64(txctx.Tx.Gas),
			GasUsed:   int64(txctx.GasUsed),
		}
	case abcitypes.CheckTxType_Recheck:
		// do nothing
	}
	return abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK}
}

func (ctrler *RigoApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	if req.Header.Height != ctrler.lastBlockCtx.Height()+1 {
		panic(fmt.Errorf("error block height: expected(%v), actural(%v)", ctrler.lastBlockCtx.Height()+1, req.Header.Height))
	}
	ctrler.logger.Debug("RigoApp::BeginBlock",
		"height", req.Header.Height,
		"hash", req.Hash,
		"prev.hash", req.Header.LastBlockId.Hash)

	ctrler.mtx.Lock() // this lock will be unlocked at EndBlock

	ctrler.nextBlockCtx = rctypes.NewBlockContext(req, ctrler.govCtrler, ctrler.acctCtrler, ctrler.stakeCtrler)

	ev0, xerr := ctrler.govCtrler.BeginBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	ev1, xerr := ctrler.stakeCtrler.BeginBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	ev2, xerr := ctrler.vmCtrler.BeginBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}

	return abcitypes.ResponseBeginBlock{
		Events: append(ev0, append(ev1, ev2...)...),
	}
}

func (ctrler *RigoApp) deliverTxSync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {

	txctx, xerr := rctypes.NewTrxContext(req.Tx,
		ctrler.nextBlockCtx.Height(),
		ctrler.nextBlockCtx.TimeSeconds(),
		true,
		func(_txctx *rctypes.TrxContext) xerrors.XError {
			_txctx.TxIdx = ctrler.nextBlockCtx.TxsCnt()
			ctrler.nextBlockCtx.AddTxsCnt(1)

			_txctx.TrxGovHandler = ctrler.govCtrler
			_txctx.TrxAcctHandler = ctrler.acctCtrler
			_txctx.TrxStakeHandler = ctrler.stakeCtrler
			_txctx.TrxEVMHandler = ctrler.vmCtrler
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.AcctHandler = ctrler.acctCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			_txctx.ChainID = ctrler.rootConfig.ChainID
			return nil
		})
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		ctrler.logger.Error("deliverTxSync", "error", xerr)
		return abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
	}
	xerr = ctrler.txExecutor.ExecuteSync(txctx)
	if xerr != nil {
		xerr = xerrors.ErrDeliverTx.Wrap(xerr)
		ctrler.logger.Error("deliverTxSync", "error", xerr)
		return abcitypes.ResponseDeliverTx{
			Code: xerr.Code(),
			Log:  xerr.Error(),
		}
	} else {

		ctrler.nextBlockCtx.AddFee(rctypes.GasToFee(txctx.GasUsed, ctrler.govCtrler.GasPrice()))

		// add event
		txctx.Events = append(txctx.Events, abcitypes.Event{
			Type: "tx",
			Attributes: []abcitypes.EventAttribute{
				{Key: []byte(rctypes.EVENT_ATTR_TXTYPE), Value: []byte(txctx.Tx.TypeString()), Index: true},
				{Key: []byte(rctypes.EVENT_ATTR_TXSENDER), Value: []byte(txctx.Tx.From.String()), Index: true},
				{Key: []byte(rctypes.EVENT_ATTR_TXRECVER), Value: []byte(txctx.Tx.To.String()), Index: true},
				{Key: []byte(rctypes.EVENT_ATTR_ADDRPAIR), Value: []byte(txctx.Tx.From.String() + txctx.Tx.To.String()), Index: true},
				{Key: []byte(rctypes.EVENT_ATTR_AMOUNT), Value: []byte(txctx.Tx.Amount.Dec()), Index: false},
			},
		})

		return abcitypes.ResponseDeliverTx{
			Code:      abcitypes.CodeTypeOK,
			GasWanted: int64(txctx.Tx.Gas),
			GasUsed:   int64(txctx.GasUsed),
			Data:      txctx.RetData,
			Events:    txctx.Events,
		}
	}
}

// deliverTxAsync is not fully implemented yet
// todo: Fully implement deliverTxAsync which processes txs in parallel.
func (ctrler *RigoApp) deliverTxAsync(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	txIdx := ctrler.nextBlockCtx.TxsCnt()
	ctrler.nextBlockCtx.AddTxsCnt(1)

	txctx, xerr := rctypes.NewTrxContext(req.Tx,
		ctrler.nextBlockCtx.Height(),
		ctrler.nextBlockCtx.TimeSeconds(),
		true,
		func(_txctx *rctypes.TrxContext) xerrors.XError {
			_txctx.TxIdx = txIdx
			_txctx.TrxGovHandler = ctrler.govCtrler
			_txctx.TrxAcctHandler = ctrler.acctCtrler
			_txctx.TrxStakeHandler = ctrler.stakeCtrler
			_txctx.TrxEVMHandler = ctrler.vmCtrler
			_txctx.GovHandler = ctrler.govCtrler
			_txctx.AcctHandler = ctrler.acctCtrler
			_txctx.StakeHandler = ctrler.stakeCtrler
			_txctx.ChainID = ctrler.rootConfig.ChainID

			// when the 'tx' is finished, it's called
			_txctx.Callback = func(ctx *rctypes.TrxContext, xerr xerrors.XError) {
				// it is called from executionRoutine goroutine
				// when execution is finished or error is generated
				response := abcitypes.ResponseDeliverTx{}
				if xerr != nil {
					xerr = xerrors.ErrDeliverTx.Wrap(xerr)
					response.Code = xerr.Code()
					response.Log = xerr.Error()

				} else {
					response.GasWanted = int64(ctx.Tx.Gas)
					response.GasUsed = int64(ctx.GasUsed)
					response.Data = ctx.RetData
					response.Events = []abcitypes.Event{
						{
							Type: "tx",
							Attributes: []abcitypes.EventAttribute{
								{Key: []byte(rctypes.EVENT_ATTR_TXTYPE), Value: []byte(ctx.Tx.TypeString()), Index: true},
								{Key: []byte(rctypes.EVENT_ATTR_TXSENDER), Value: []byte(ctx.Tx.From.String()), Index: true},
								{Key: []byte(rctypes.EVENT_ATTR_TXRECVER), Value: []byte(ctx.Tx.To.String()), Index: true},
								{Key: []byte(rctypes.EVENT_ATTR_ADDRPAIR), Value: []byte(ctx.Tx.From.String() + ctx.Tx.To.String()), Index: true},
							},
						},
					}

					ctrler.nextBlockCtx.AddFee(rctypes.GasToFee(ctx.GasUsed, ctrler.govCtrler.GasPrice()))
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
	return ctrler.deliverTxSync(req)
}

func (ctrler *RigoApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	ctrler.logger.Debug("Begin RigoApp::EndBlock",
		"height", req.Height)

	defer func() {
		ctrler.mtx.Unlock() // this was locked at BeginBlock
		ctrler.logger.Debug("Finish RigoApp::EndBlock",
			"height", req.Height)
	}()

	ev0, xerr := ctrler.govCtrler.EndBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	ev1, xerr := ctrler.acctCtrler.EndBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	ev2, xerr := ctrler.stakeCtrler.EndBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}
	ev3, xerr := ctrler.vmCtrler.EndBlock(ctrler.nextBlockCtx)
	if xerr != nil {
		ctrler.logger.Error("RigoApp", "error", xerr)
		panic(xerr)
	}

	var ev []abcitypes.Event
	ev = append(ev, ev0...)
	ev = append(ev, ev1...)
	ev = append(ev, ev2...)
	ev = append(ev, ev3...)

	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: ctrler.nextBlockCtx.ValUpdates,
		Events:           ev,
	}
}

func (ctrler *RigoApp) Commit() abcitypes.ResponseCommit {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.logger.Debug("RigoApp::Commit", "height", ctrler.nextBlockCtx.Height())

	appHash0, ver0, err := ctrler.govCtrler.Commit()
	if err != nil {
		panic(err)
	}
	//ctrler.logger.Debug("RigoApp::Commit", "height", ver0, "appHash0", bytes.HexBytes(appHash0))

	appHash1, ver1, err := ctrler.acctCtrler.Commit()
	if err != nil {
		panic(err)
	}
	//ctrler.logger.Debug("RigoApp::Commit", "height", ver1, "appHash1", bytes.HexBytes(appHash1))

	appHash2, ver2, err := ctrler.stakeCtrler.Commit()
	if err != nil {
		panic(err)
	}
	//ctrler.logger.Debug("RigoApp::Commit", "height", ver2, "appHash2", bytes.HexBytes(appHash2))

	appHash3, ver3, err := ctrler.vmCtrler.Commit()
	if err != nil {
		panic(err)
	}
	//ctrler.logger.Debug("RigoApp::Commit", "height", ver3, "appHash3", bytes.HexBytes(appHash3))

	if ver0 != ver1 || ver1 != ver2 || ver2 != ver3 {
		panic(fmt.Sprintf("Not same versions: gov: %v, account:%v, stake:%v, vm:%v", ver0, ver1, ver2, ver3))
	}

	appHash := crypto.DefaultHash(appHash0, appHash1, appHash2, appHash3)
	ctrler.nextBlockCtx.SetAppHash(appHash)
	ctrler.logger.Debug("RigoApp::Commit", "height", ver0, "txs", ctrler.nextBlockCtx.TxsCnt(), "app hash", ctrler.nextBlockCtx.AppHash())

	ctrler.metaDB.PutLastBlockContext(ctrler.nextBlockCtx)
	ctrler.metaDB.PutLastBlockHeight(ver0)

	ctrler.lastBlockCtx = ctrler.nextBlockCtx
	ctrler.nextBlockCtx = nil

	return abcitypes.ResponseCommit{
		Data: appHash[:],
	}
}
