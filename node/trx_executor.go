package node

import (
	"bytes"
	"fmt"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
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
	if len(ctx.Tx.From) != types.AddrSize {
		return xerrors.ErrInvalidAddress
	}

	if len(ctx.Tx.To) != types.AddrSize {
		return xerrors.ErrInvalidAddress
	}

	// check signature
	var fromAddr types.Address
	var pubBytes abytes.HexBytes

	if ctx.Exec {
		tx := ctx.Tx
		sig := tx.Sig
		tx.Sig = nil
		_txbz, xerr := tx.Encode()
		tx.Sig = sig
		if xerr != nil {
			return xerr
		}
		if fromAddr, pubBytes, xerr = crypto.Sig2Addr(_txbz, sig); xerr != nil {
			return xerr
		}
		if bytes.Compare(fromAddr, tx.From) != 0 {
			return xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, fromAddr))
		}
		ctx.SenderPubKey = pubBytes
	}

	//
	// tx validation
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

func runTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {

	//
	// tx execution
	if ctx.Tx.GetType() == ctrlertypes.TRX_CONTRACT {
		if xerr := ctx.TrxEVMHandler.ExecuteTrx(ctx); xerr != nil && xerr != xerrors.ErrUnknownTrxType {
			return xerr
		}
	} else {
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
	}

	return nil
}
