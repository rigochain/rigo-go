package node

import (
	"github.com/kysee/arcanus/ctrlers/types"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
)

type TrxExecutor struct {
	txCtxChs []chan *types.TrxContext
	logger   log.Logger
}

func NewTrxExecutor(n int, logger log.Logger) *TrxExecutor {
	txCtxChs := make([]chan *types.TrxContext, n)
	for i := 0; i < n; i++ {
		txCtxChs[i] = make(chan *types.TrxContext, 128)
	}
	return &TrxExecutor{
		txCtxChs: txCtxChs,
		logger:   logger,
	}
}

func (txe *TrxExecutor) Start() {
	for _, ch := range txe.txCtxChs {
		go executionRoutine(ch, txe.logger)
	}
}

func (txe *TrxExecutor) Stop() {
	for _, ch := range txe.txCtxChs {
		close(ch)
	}
	txe.txCtxChs = nil
}

func (txe *TrxExecutor) ExecuteSync(ctx *types.TrxContext) xerrors.XError {
	return runTrx(ctx)
}

func (txe *TrxExecutor) ExecuteAsync(ctx *types.TrxContext) xerrors.XError {
	n := len(txe.txCtxChs)
	i := int(ctx.Tx.From[0]) % n

	if txe.txCtxChs[i] == nil {
		return xerrors.New("channel is not available")
	}
	txe.txCtxChs[i] <- ctx
	return nil
}

func executionRoutine(ch chan *types.TrxContext, logger log.Logger) {
	logger.Info("Start transaction execution routine")

	for ctx := range ch {
		ctx.Callback(ctx, runTrx(ctx))
	}
}

func runTrx(ctx *types.TrxContext) xerrors.XError {

	//
	// tx validation
	if xerr := ctx.GovHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.AcctHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.StakeHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}

	//
	// tx execution
	if xerr := ctx.GovHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.AcctHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.StakeHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}

	return nil
}
