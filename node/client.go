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
	l.app.(*ArcanusApp).SetLocalClient(client)
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

	deliverTxReqReses []*abcicli.ReqRes
	OnTrxExecFinished func(abcicli.Client, int, *abcitypes.RequestDeliverTx, *abcitypes.ResponseDeliverTx)
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
		mtx:               mtx,
		Application:       app,
		OnTrxExecFinished: defualtOnTrxExecFinished,
	}
	cli.BaseService = *service.NewBaseService(nil, "arcanusLocalClient", cli)
	return cli
}

func (client *arcanusLocalClient) OnStart() error {
	return client.Application.(*ArcanusApp).Start()
}

func (client *arcanusLocalClient) OnStop() {
	client.Application.(*ArcanusApp).Stop()
}

func (client *arcanusLocalClient) SetResponseCallback(cb abcicli.Callback) {
	client.mtx.Lock()
	client.Callback = cb
	client.mtx.Unlock()
}

// TODO: change abcitypes.Application to include Error()?
func (client *arcanusLocalClient) Error() error {
	return nil
}

func (client *arcanusLocalClient) FlushAsync() *abcicli.ReqRes {
	// Do nothing
	return newLocalReqRes(abcitypes.ToRequestFlush(), nil)
}

func (client *arcanusLocalClient) EchoAsync(msg string) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	return client.callback(
		abcitypes.ToRequestEcho(msg),
		abcitypes.ToResponseEcho(msg),
	)
}

func (client *arcanusLocalClient) InfoAsync(req abcitypes.RequestInfo) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Info(req)
	return client.callback(
		abcitypes.ToRequestInfo(req),
		abcitypes.ToResponseInfo(res),
	)
}

func (client *arcanusLocalClient) SetOptionAsync(req abcitypes.RequestSetOption) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.SetOption(req)
	return client.callback(
		abcitypes.ToRequestSetOption(req),
		abcitypes.ToResponseSetOption(res),
	)
}

func (client *arcanusLocalClient) DeliverTxAsync(params abcitypes.RequestDeliverTx) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	rr := newLocalReqRes(abcitypes.ToRequestDeliverTx(params), abcitypes.ToResponseDeliverTx(abcitypes.ResponseDeliverTx{}))
	client.deliverTxReqReses = append(client.deliverTxReqReses, rr)

	_ = client.Application.DeliverTx(params)
	return nil

	//res := client.Application.DeliverTx(params)
	//return client.callback(
	//	abcitypes.ToRequestDeliverTx(params),
	//	abcitypes.ToResponseDeliverTx(res),
	//)
}

func defualtOnTrxExecFinished(client abcicli.Client, txidx int, req *abcitypes.RequestDeliverTx, res *abcitypes.ResponseDeliverTx) {
	client.(*arcanusLocalClient).onTrxFinished(txidx, req, res)
}

func (client *arcanusLocalClient) onTrxFinished(txidx int, req *abcitypes.RequestDeliverTx, res *abcitypes.ResponseDeliverTx) {
	if txidx >= len(client.deliverTxReqReses) {
		panic("client.deliverTxReqReses's length is wrong")
	}
	rr := client.deliverTxReqReses[txidx]
	rr.Response = abcitypes.ToResponseDeliverTx(*res)
	rr.Done()
}

func (client *arcanusLocalClient) CheckTxAsync(req abcitypes.RequestCheckTx) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.CheckTx(req)
	return client.callback(
		abcitypes.ToRequestCheckTx(req),
		abcitypes.ToResponseCheckTx(res),
	)
}

func (client *arcanusLocalClient) QueryAsync(req abcitypes.RequestQuery) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Query(req)
	return client.callback(
		abcitypes.ToRequestQuery(req),
		abcitypes.ToResponseQuery(res),
	)
}

func (client *arcanusLocalClient) CommitAsync() *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Commit()
	return client.callback(
		abcitypes.ToRequestCommit(),
		abcitypes.ToResponseCommit(res),
	)
}

func (client *arcanusLocalClient) InitChainAsync(req abcitypes.RequestInitChain) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.InitChain(req)
	return client.callback(
		abcitypes.ToRequestInitChain(req),
		abcitypes.ToResponseInitChain(res),
	)
}

func (client *arcanusLocalClient) BeginBlockAsync(req abcitypes.RequestBeginBlock) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.BeginBlock(req)
	return client.callback(
		abcitypes.ToRequestBeginBlock(req),
		abcitypes.ToResponseBeginBlock(res),
	)
}

func (client *arcanusLocalClient) EndBlockAsync(req abcitypes.RequestEndBlock) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.EndBlock(req)
	return client.callback(
		abcitypes.ToRequestEndBlock(req),
		abcitypes.ToResponseEndBlock(res),
	)
}

func (client *arcanusLocalClient) ListSnapshotsAsync(req abcitypes.RequestListSnapshots) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ListSnapshots(req)
	return client.callback(
		abcitypes.ToRequestListSnapshots(req),
		abcitypes.ToResponseListSnapshots(res),
	)
}

func (client *arcanusLocalClient) OfferSnapshotAsync(req abcitypes.RequestOfferSnapshot) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.OfferSnapshot(req)
	return client.callback(
		abcitypes.ToRequestOfferSnapshot(req),
		abcitypes.ToResponseOfferSnapshot(res),
	)
}

func (client *arcanusLocalClient) LoadSnapshotChunkAsync(req abcitypes.RequestLoadSnapshotChunk) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.LoadSnapshotChunk(req)
	return client.callback(
		abcitypes.ToRequestLoadSnapshotChunk(req),
		abcitypes.ToResponseLoadSnapshotChunk(res),
	)
}

func (client *arcanusLocalClient) ApplySnapshotChunkAsync(req abcitypes.RequestApplySnapshotChunk) *abcicli.ReqRes {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ApplySnapshotChunk(req)
	return client.callback(
		abcitypes.ToRequestApplySnapshotChunk(req),
		abcitypes.ToResponseApplySnapshotChunk(res),
	)
}

//-------------------------------------------------------

func (client *arcanusLocalClient) FlushSync() error {
	return nil
}

func (client *arcanusLocalClient) EchoSync(msg string) (*abcitypes.ResponseEcho, error) {
	return &abcitypes.ResponseEcho{Message: msg}, nil
}

func (client *arcanusLocalClient) InfoSync(req abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Info(req)
	return &res, nil
}

func (client *arcanusLocalClient) SetOptionSync(req abcitypes.RequestSetOption) (*abcitypes.ResponseSetOption, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.SetOption(req)
	return &res, nil
}

func (client *arcanusLocalClient) DeliverTxSync(req abcitypes.RequestDeliverTx) (*abcitypes.ResponseDeliverTx, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.DeliverTx(req)
	return &res, nil
}

func (client *arcanusLocalClient) CheckTxSync(req abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.CheckTx(req)
	return &res, nil
}

func (client *arcanusLocalClient) QuerySync(req abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Query(req)
	return &res, nil
}

func (client *arcanusLocalClient) CommitSync() (*abcitypes.ResponseCommit, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.Commit()
	return &res, nil
}

func (client *arcanusLocalClient) InitChainSync(req abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.InitChain(req)
	return &res, nil
}

func (client *arcanusLocalClient) BeginBlockSync(req abcitypes.RequestBeginBlock) (*abcitypes.ResponseBeginBlock, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.BeginBlock(req)
	return &res, nil
}

func (client *arcanusLocalClient) EndBlockSync(req abcitypes.RequestEndBlock) (*abcitypes.ResponseEndBlock, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	// wait that all txs are finished
	for _, rr := range client.deliverTxReqReses {
		rr.Wait()
		client.Callback(rr.Request, rr.Response)
		rr.InvokeCallback()
	}
	client.deliverTxReqReses = nil

	res := client.Application.EndBlock(req)
	return &res, nil
}

func (client *arcanusLocalClient) ListSnapshotsSync(req abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ListSnapshots(req)
	return &res, nil
}

func (client *arcanusLocalClient) OfferSnapshotSync(req abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.OfferSnapshot(req)
	return &res, nil
}

func (client *arcanusLocalClient) LoadSnapshotChunkSync(
	req abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (client *arcanusLocalClient) ApplySnapshotChunkSync(
	req abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	res := client.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (client *arcanusLocalClient) callback(req *abcitypes.Request, res *abcitypes.Response) *abcicli.ReqRes {
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
