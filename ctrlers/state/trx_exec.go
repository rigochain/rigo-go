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
	return nil
}
