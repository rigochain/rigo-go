package state

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	tmproxy "github.com/tendermint/tendermint/proxy"
)

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type arcanusLocalClientCreator struct {
	mtx *tmsync.Mutex
	app abcitypes.Application
}

// NewLocalClientCreator returns a ClientCreator for the given app,
// which will be running locally.
func NewArcanusLocalClientCreator(app abcitypes.Application) tmproxy.ClientCreator {
	return &arcanusLocalClientCreator{
		mtx: new(tmsync.Mutex),
		app: app,
	}
}

func (l *arcanusLocalClientCreator) NewABCIClient() (abcicli.Client, error) {
	client := NewArcanusLocalClient(l.mtx, l.app)
	l.app.(*ChainCtrler).SetLocalClient(client)
	return client, nil
}

var _ abcicli.Client = (*arcanusLocalClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type arcanusLocalClient struct {
	service.BaseService

	mtx *tmsync.Mutex
	abcitypes.Application
	abcicli.Callback
}

var _ abcicli.Client = (*arcanusLocalClient)(nil)

// NewLocalClient creates a local client, which will be directly calling the
// methods of the given app.
//
// Both Async and Sync methods ignore the given context.Context parameter.
func NewArcanusLocalClient(mtx *tmsync.Mutex, app abcitypes.Application) abcicli.Client {
	if mtx == nil {
		mtx = new(tmsync.Mutex)
	}
	cli := &arcanusLocalClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "arcanusLocalClient", cli)
	return cli
}

func (app *arcanusLocalClient) SetResponseCallback(cb abcicli.Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// TODO: change abcitypes.Application to include Error()?
func (app *arcanusLocalClient) Error() error {
	return nil
}

func (app *arcanusLocalClient) FlushAsync() *abcicli.ReqRes {
	// Do nothing
	return newLocalReqRes(abcitypes.ToRequestFlush(), nil)
}

func (app *arcanusLocalClient) EchoAsync(msg string) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.callback(
		abcitypes.ToRequestEcho(msg),
		abcitypes.ToResponseEcho(msg),
	)
}

func (app *arcanusLocalClient) InfoAsync(req abcitypes.RequestInfo) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return app.callback(
		abcitypes.ToRequestInfo(req),
		abcitypes.ToResponseInfo(res),
	)
}

func (app *arcanusLocalClient) SetOptionAsync(req abcitypes.RequestSetOption) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return app.callback(
		abcitypes.ToRequestSetOption(req),
		abcitypes.ToResponseSetOption(res),
	)
}

func (app *arcanusLocalClient) DeliverTxAsync(params abcitypes.RequestDeliverTx) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(params)
	return app.callback(
		abcitypes.ToRequestDeliverTx(params),
		abcitypes.ToResponseDeliverTx(res),
	)
}

func (app *arcanusLocalClient) CheckTxAsync(req abcitypes.RequestCheckTx) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return app.callback(
		abcitypes.ToRequestCheckTx(req),
		abcitypes.ToResponseCheckTx(res),
	)
}

func (app *arcanusLocalClient) QueryAsync(req abcitypes.RequestQuery) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return app.callback(
		abcitypes.ToRequestQuery(req),
		abcitypes.ToResponseQuery(res),
	)
}

func (app *arcanusLocalClient) CommitAsync() *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return app.callback(
		abcitypes.ToRequestCommit(),
		abcitypes.ToResponseCommit(res),
	)
}

func (app *arcanusLocalClient) InitChainAsync(req abcitypes.RequestInitChain) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return app.callback(
		abcitypes.ToRequestInitChain(req),
		abcitypes.ToResponseInitChain(res),
	)
}

func (app *arcanusLocalClient) BeginBlockAsync(req abcitypes.RequestBeginBlock) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return app.callback(
		abcitypes.ToRequestBeginBlock(req),
		abcitypes.ToResponseBeginBlock(res),
	)
}

func (app *arcanusLocalClient) EndBlockAsync(req abcitypes.RequestEndBlock) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return app.callback(
		abcitypes.ToRequestEndBlock(req),
		abcitypes.ToResponseEndBlock(res),
	)
}

func (app *arcanusLocalClient) ListSnapshotsAsync(req abcitypes.RequestListSnapshots) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ListSnapshots(req)
	return app.callback(
		abcitypes.ToRequestListSnapshots(req),
		abcitypes.ToResponseListSnapshots(res),
	)
}

func (app *arcanusLocalClient) OfferSnapshotAsync(req abcitypes.RequestOfferSnapshot) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.OfferSnapshot(req)
	return app.callback(
		abcitypes.ToRequestOfferSnapshot(req),
		abcitypes.ToResponseOfferSnapshot(res),
	)
}

func (app *arcanusLocalClient) LoadSnapshotChunkAsync(req abcitypes.RequestLoadSnapshotChunk) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.LoadSnapshotChunk(req)
	return app.callback(
		abcitypes.ToRequestLoadSnapshotChunk(req),
		abcitypes.ToResponseLoadSnapshotChunk(res),
	)
}

func (app *arcanusLocalClient) ApplySnapshotChunkAsync(req abcitypes.RequestApplySnapshotChunk) *abcicli.ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return app.callback(
		abcitypes.ToRequestApplySnapshotChunk(req),
		abcitypes.ToResponseApplySnapshotChunk(res),
	)
}

//-------------------------------------------------------

func (app *arcanusLocalClient) FlushSync() error {
	return nil
}

func (app *arcanusLocalClient) EchoSync(msg string) (*abcitypes.ResponseEcho, error) {
	return &abcitypes.ResponseEcho{Message: msg}, nil
}

func (app *arcanusLocalClient) InfoSync(req abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return &res, nil
}

func (app *arcanusLocalClient) SetOptionSync(req abcitypes.RequestSetOption) (*abcitypes.ResponseSetOption, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return &res, nil
}

func (app *arcanusLocalClient) DeliverTxSync(req abcitypes.RequestDeliverTx) (*abcitypes.ResponseDeliverTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(req)
	return &res, nil
}

func (app *arcanusLocalClient) CheckTxSync(req abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return &res, nil
}

func (app *arcanusLocalClient) QuerySync(req abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return &res, nil
}

func (app *arcanusLocalClient) CommitSync() (*abcitypes.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return &res, nil
}

func (app *arcanusLocalClient) InitChainSync(req abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return &res, nil
}

func (app *arcanusLocalClient) BeginBlockSync(req abcitypes.RequestBeginBlock) (*abcitypes.ResponseBeginBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return &res, nil
}

func (app *arcanusLocalClient) EndBlockSync(req abcitypes.RequestEndBlock) (*abcitypes.ResponseEndBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return &res, nil
}

func (app *arcanusLocalClient) ListSnapshotsSync(req abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ListSnapshots(req)
	return &res, nil
}

func (app *arcanusLocalClient) OfferSnapshotSync(req abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.OfferSnapshot(req)
	return &res, nil
}

func (app *arcanusLocalClient) LoadSnapshotChunkSync(
	req abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (app *arcanusLocalClient) ApplySnapshotChunkSync(
	req abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (app *arcanusLocalClient) callback(req *abcitypes.Request, res *abcitypes.Response) *abcicli.ReqRes {
	app.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.InvokeCallback() //rr.callbackInvoked = true
	return rr
}

func newLocalReqRes(req *abcitypes.Request, res *abcitypes.Response) *abcicli.ReqRes {
	reqRes := abcicli.NewReqRes(req)
	reqRes.Response = res
	return reqRes
}
