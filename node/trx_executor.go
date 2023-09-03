package node

import (
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	"runtime"
	"strconv"
	"strings"
)

type TrxExecutor struct {
	txCtxChs []chan *ctrlertypes.TrxContext
	logger   log.Logger
}

func NewTrxExecutor(n int, logger log.Logger) *TrxExecutor {
	txCtxChs := make([]chan *ctrlertypes.TrxContext, n)
	for i := 0; i < n; i++ {
		txCtxChs[i] = make(chan *ctrlertypes.TrxContext, 5000)
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

func (txe *TrxExecutor) ExecuteSync(ctx *ctrlertypes.TrxContext) xerrors.XError {
	xerr := validateTrx(ctx)
	if xerr != nil {
		return xerr
	}
	xerr = runTrx(ctx)
	if xerr != nil {
		return xerr
	}
	return nil
}

func (txe *TrxExecutor) ExecuteAsync(ctx *ctrlertypes.TrxContext) xerrors.XError {
	n := len(txe.txCtxChs)
	i := int(ctx.Tx.From[0]) % n

	if txe.txCtxChs[i] == nil {
		return xerrors.NewOrdinary("transaction execution channel is not available")
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

func executionRoutine(name string, ch chan *ctrlertypes.TrxContext, logger log.Logger) {
	logger.Info("Start transaction execution routine", "goid", goid(), "name", name)

	for ctx := range ch {
		//if ctx.Exec {
		//	logger.Info("[DEBUG] Begin of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}
		var xerr xerrors.XError

		if xerr = validateTrx(ctx); xerr == nil {
			xerr = runTrx(ctx)
		}

		//if ctx.Exec {
		//	logger.Info("[DEBUG] End of executionRoutine", "txhash", ctx.TxHash, "goid", goid(), "name", name)
		//}

		ctx.Callback(ctx, xerr)
	}
}

func validateTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// tx validation
	if xerr := commonValidation(ctx); xerr != nil {
		return xerr
	}
	if xerr := ctx.TrxGovHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.TrxAcctHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.TrxStakeHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}
	if xerr := ctx.TrxEVMHandler.ValidateTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
		return xerr
	}

	return nil
}

func commonValidation(ctx *ctrlertypes.TrxContext) xerrors.XError {
	return nil
}

func runTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// tx execution
	switch ctx.Tx.GetType() {
	case ctrlertypes.TRX_CONTRACT:
		if xerr := ctx.TrxEVMHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			return xerr
		}
	default:
		if xerr := ctx.TrxGovHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			return xerr
		}
		if xerr := ctx.TrxAcctHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			// todo: rollback changes in TrxGovHandler.ExecuteTrx
			return xerr
		}
		if xerr := ctx.TrxStakeHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			// todo: rollback changes in TrxGovHandler.ExecuteTrx and TrxAcctHandler.ExecuteTrx
			return xerr
		}
		if xerr := runPostTrx(ctx); xerr != nil {
			return xerr
		}
	}

	// processing gas

	return nil
}

func runPostTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	fee := new(uint256.Int).Mul(ctx.Tx.GasPrice, uint256.NewInt(uint64(ctx.Tx.Gas)))
	if xerr := ctx.Sender.SubBalance(fee); xerr != nil {
		return xerr
	}
	ctx.Sender.AddNonce()

	if xerr := ctx.AcctHandler.SetAccountCommittable(ctx.Sender, ctx.Exec); xerr != nil {
		return xerr
	}

	// set used gas
	ctx.GasUsed = ctx.Tx.Gas
	return nil
}
