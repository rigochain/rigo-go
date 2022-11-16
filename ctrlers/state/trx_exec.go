package state

import (
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
)

type TrxExecutor struct {
	handlerMap map[int32][]trxs.ITrxHandler
	logger     log.Logger
}

func NewTrxExecutor(handlers map[int32][]trxs.ITrxHandler, logger log.Logger) *TrxExecutor {
	return &TrxExecutor{
		handlerMap: handlers,
		logger:     logger,
	}
}

func (txe *TrxExecutor) Execute(ctx *trxs.TrxContext) error {
	// check sender account nonce
	if xerr := ctx.Sender.CheckNonce(ctx.Tx.Nonce); xerr != nil {
		return xerr
	}

	// check sender account balance
	if xerr := ctx.Sender.CheckBalance(ctx.NeedAmt); xerr != nil {
		return xerr
	}

	if handlers, ok := txe.handlerMap[ctx.Tx.Type]; !ok {
		return xerrors.ErrInvalidTrxType
	} else {
		for _, h := range handlers {
			if err := h.Validate(ctx); err != nil {
				return err
			} else if err := h.Execute(ctx); err != nil {
				return err
			}
		}
	}

	if err := ctx.Sender.SubBalance(ctx.Tx.Gas); err != nil {
		return err
	}
	ctx.GasUsed = ctx.Tx.Gas
	ctx.Sender.AddNonce()

	return nil
}
