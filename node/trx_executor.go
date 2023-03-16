package node

import (
	"fmt"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	"runtime"
	"strconv"
	"strings"
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
	for i, ch := range txe.txCtxChs {
		go executionRoutine(fmt.Sprintf("executionRoutine-%d", i), ch, txe.logger)
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
	//if ctx.Exec {
	//	txe.logger.Info("[DEBUG] TrxExecutor::ExecuteAsync", "index", i, "txhash", ctx.TxHash)
	//}
	txe.txCtxChs[i] <- ctx
	return nil
}

// for test
func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func executionRoutine(name string, ch chan *types.TrxContext, logger log.Logger) {
	logger.Info("Start transaction execution routine", "goid", goid(), "name", name)

	for ctx := range ch {
		//if ctx.Exec {
		//	logger.Info("[DEBUG] Begin of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}

		xerr := runTrx(ctx)

		//if ctx.Exec {
		//	logger.Info("[DEBUG] End of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}

		ctx.Callback(ctx, xerr)
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
		// todo: rollback changes in GovHandler.ExecuteTrx
		return xerr
	}
	if xerr := ctx.StakeHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		// todo: rollback changes in GovHandler.ExecuteTrx and AcctHandler.ExecuteTrx
		return xerr
	}

	return nil
}
