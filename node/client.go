package node

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	tmproxy "github.com/tendermint/tendermint/proxy"
)

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type rigoLocalClientCreator struct {
	mtx *tmsync.Mutex
	app abcitypes.Application
}

// NewLocalClientCreator returns a ClientCreator for the given app,
// which will be running locally.
func NewRigoLocalClientCreator(app abcitypes.Application) tmproxy.ClientCreator {
	return &rigoLocalClientCreator{
		mtx: new(tmsync.Mutex),
		app: app,
	}
}

func (l *rigoLocalClientCreator) NewABCIClient() (abcicli.Client, error) {
	client := NewRigoLocalClient(l.mtx, l.app)
	l.app.(*RigoApp).SetLocalClient(client)
	return client, nil
}

var _ abcicli.Client = (*rigoLocalClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type rigoLocalClient struct {
	service.BaseService

	mtx *tmsync.Mutex
	abcitypes.Application
	abcicli.Callback

	deliverTxReqReses []*abcicli.ReqRes
	OnTrxExecFinished func(abcicli.Client, int, *abcitypes.RequestDeliverTx, *abcitypes.ResponseDeliverTx)
}

var _ abcicli.Client = (*rigoLocalClient)(nil)

// NewLocalClient creates a local web3, which will be directly calling the
// methods of the given app.
//
// Both Async and Sync methods ignore the given context.Context parameter.
func NewRigoLocalClient(mtx *tmsync.Mutex, app abcitypes.Application) abcicli.Client {
	if mtx == nil {
		mtx = new(tmsync.Mutex)
	}
	cli := &rigoLocalClient{
		mtx:               mtx,
		Application:       app,
		OnTrxExecFinished: defualtOnTrxExecFinished,
	}
	cli.BaseService = *service.NewBaseService(nil, "rigoLocalClient", cli)
	return cli
}

func (client *rigoLocalClient) OnStart() error {
	return client.Application.(*RigoApp).Start()
}

func (client *rigoLocalClient) OnStop() {
	client.Application.(*RigoApp).Stop()
}

func (client *rigoLocalClient) SetResponseCallback(cb abcicli.Callback) {
	client.mtx.Lock()
	client.Callback = cb
	client.mtx.Unlock()
}

// TODO: change abcitypes.Application to include Error()?
func (client *rigoLocalClient) Error() error {
	return nil
}

func (client *rigoLocalClient) FlushAsync() *abcicli.ReqRes {
	// Do nothing
	return newLocalReqRes(abcitypes.ToRequestFlush(), nil)
}

func (client *rigoLocalClient) EchoAsync(msg string) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	return client.callback(
		abcitypes.ToRequestEcho(msg),
		abcitypes.ToResponseEcho(msg),
	)
}

func (client *rigoLocalClient) InfoAsync(req abcitypes.RequestInfo) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Info(req)
	return client.callback(
		abcitypes.ToRequestInfo(req),
		abcitypes.ToResponseInfo(res),
	)
}

func (client *rigoLocalClient) SetOptionAsync(req abcitypes.RequestSetOption) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.SetOption(req)
	return client.callback(
		abcitypes.ToRequestSetOption(req),
		abcitypes.ToResponseSetOption(res),
	)
}

func (client *rigoLocalClient) DeliverTxAsync(params abcitypes.RequestDeliverTx) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	// todo: Implement parallel tx processing
	//rr := newLocalReqRes(abcitypes.ToRequestDeliverTx(params), abcitypes.ToResponseDeliverTx(abcitypes.ResponseDeliverTx{}))
	//client.deliverTxReqReses = append(client.deliverTxReqReses, rr)
	//
	//_ = client.Application.DeliverTx(params)
	//return nil

	res := client.Application.DeliverTx(params)
	return client.callback(
		abcitypes.ToRequestDeliverTx(params),
		abcitypes.ToResponseDeliverTx(res),
	)
}

func defualtOnTrxExecFinished(client abcicli.Client, txidx int, req *abcitypes.RequestDeliverTx, res *abcitypes.ResponseDeliverTx) {
	client.(*rigoLocalClient).onTrxFinished(txidx, req, res)
}

func (client *rigoLocalClient) onTrxFinished(txidx int, req *abcitypes.RequestDeliverTx, res *abcitypes.ResponseDeliverTx) {
	client.Logger.Info("rigoLocalClient receives finished tx", "index", txidx)
	if txidx >= len(client.deliverTxReqReses) {
		panic("web3.deliverTxReqReses's length is wrong")
	}
	rr := client.deliverTxReqReses[txidx]
	rr.Response = abcitypes.ToResponseDeliverTx(*res)
	rr.Done()
}

func (client *rigoLocalClient) CheckTxAsync(req abcitypes.RequestCheckTx) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.CheckTx(req)
	return client.callback(
		abcitypes.ToRequestCheckTx(req),
		abcitypes.ToResponseCheckTx(res),
	)
}

func (client *rigoLocalClient) QueryAsync(req abcitypes.RequestQuery) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Query(req)
	return client.callback(
		abcitypes.ToRequestQuery(req),
		abcitypes.ToResponseQuery(res),
	)
}

func (client *rigoLocalClient) CommitAsync() *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Commit()
	return client.callback(
		abcitypes.ToRequestCommit(),
		abcitypes.ToResponseCommit(res),
	)
}

func (client *rigoLocalClient) InitChainAsync(req abcitypes.RequestInitChain) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.InitChain(req)
	return client.callback(
		abcitypes.ToRequestInitChain(req),
		abcitypes.ToResponseInitChain(res),
	)
}

func (client *rigoLocalClient) BeginBlockAsync(req abcitypes.RequestBeginBlock) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.BeginBlock(req)
	return client.callback(
		abcitypes.ToRequestBeginBlock(req),
		abcitypes.ToResponseBeginBlock(res),
	)
}

func (client *rigoLocalClient) EndBlockAsync(req abcitypes.RequestEndBlock) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.EndBlock(req)
	return client.callback(
		abcitypes.ToRequestEndBlock(req),
		abcitypes.ToResponseEndBlock(res),
	)
}

func (client *rigoLocalClient) ListSnapshotsAsync(req abcitypes.RequestListSnapshots) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ListSnapshots(req)
	return client.callback(
		abcitypes.ToRequestListSnapshots(req),
		abcitypes.ToResponseListSnapshots(res),
	)
}

func (client *rigoLocalClient) OfferSnapshotAsync(req abcitypes.RequestOfferSnapshot) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.OfferSnapshot(req)
	return client.callback(
		abcitypes.ToRequestOfferSnapshot(req),
		abcitypes.ToResponseOfferSnapshot(res),
	)
}

func (client *rigoLocalClient) LoadSnapshotChunkAsync(req abcitypes.RequestLoadSnapshotChunk) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.LoadSnapshotChunk(req)
	return client.callback(
		abcitypes.ToRequestLoadSnapshotChunk(req),
		abcitypes.ToResponseLoadSnapshotChunk(res),
	)
}

func (client *rigoLocalClient) ApplySnapshotChunkAsync(req abcitypes.RequestApplySnapshotChunk) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ApplySnapshotChunk(req)
	return client.callback(
		abcitypes.ToRequestApplySnapshotChunk(req),
		abcitypes.ToResponseApplySnapshotChunk(res),
	)
}

//-------------------------------------------------------

func (client *rigoLocalClient) FlushSync() error {
	return nil
}

func (client *rigoLocalClient) EchoSync(msg string) (*abcitypes.ResponseEcho, error) {
	return &abcitypes.ResponseEcho{Message: msg}, nil
}

func (client *rigoLocalClient) InfoSync(req abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Info(req)
	return &res, nil
}

func (client *rigoLocalClient) SetOptionSync(req abcitypes.RequestSetOption) (*abcitypes.ResponseSetOption, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.SetOption(req)
	return &res, nil
}

func (client *rigoLocalClient) DeliverTxSync(req abcitypes.RequestDeliverTx) (*abcitypes.ResponseDeliverTx, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.DeliverTx(req)
	return &res, nil
}

func (client *rigoLocalClient) CheckTxSync(req abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.CheckTx(req)
	return &res, nil
}

func (client *rigoLocalClient) QuerySync(req abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Query(req)
	return &res, nil
}

func (client *rigoLocalClient) CommitSync() (*abcitypes.ResponseCommit, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Commit()
	return &res, nil
}

func (client *rigoLocalClient) InitChainSync(req abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.InitChain(req)
	return &res, nil
}

func (client *rigoLocalClient) BeginBlockSync(req abcitypes.RequestBeginBlock) (*abcitypes.ResponseBeginBlock, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.BeginBlock(req)
	return &res, nil
}

func (client *rigoLocalClient) EndBlockSync(req abcitypes.RequestEndBlock) (*abcitypes.ResponseEndBlock, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	// todo: Implement parallel tx processing
	//// wait that all txs are finished
	//for _, rr := range client.deliverTxReqReses {
	//	rr.Wait()
	//	client.Callback(rr.Request, rr.Response)
	//	rr.InvokeCallback()
	//}
	//client.deliverTxReqReses = nil

	res := client.Application.EndBlock(req)
	return &res, nil
}

func (client *rigoLocalClient) ListSnapshotsSync(req abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ListSnapshots(req)
	return &res, nil
}

func (client *rigoLocalClient) OfferSnapshotSync(req abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.OfferSnapshot(req)
	return &res, nil
}

func (client *rigoLocalClient) LoadSnapshotChunkSync(
	req abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (client *rigoLocalClient) ApplySnapshotChunkSync(
	req abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (client *rigoLocalClient) callback(req *abcitypes.Request, res *abcitypes.Response) *abcicli.ReqRes {
	client.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.InvokeCallback() //rr.callbackInvoked = true
	return rr
}

func newLocalReqRes(req *abcitypes.Request, res *abcitypes.Response) *abcicli.ReqRes {
	reqRes := abcicli.NewReqRes(req)
	reqRes.Response = res
	return reqRes
}
