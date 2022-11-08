package state

import (
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
)

type TrxExecutor struct {
	handlers []trxs.ITrxHandler
	logger   log.Logger
}

func NewTrxExecutor(logger log.Logger, handlers ...trxs.ITrxHandler) *TrxExecutor {
	return &TrxExecutor{
		handlers: handlers,
		logger:   logger,
	}
}

func (txe *TrxExecutor) Execute(ctx *trxs.TrxContext) error {
	switch ctx.Tx.Type {
	case trxs.TRX_TRANSFER:
		h := txe.handlers[0]
		if err := h.Validate(ctx); err != nil {
			return err
		} else if err := h.Apply(ctx); err != nil {
			return err
		}
	case trxs.TRX_STAKING, trxs.TRX_UNSTAKING:
		h := txe.handlers[1]
		if err := h.Validate(ctx); err != nil {
			return err
		} else if err := h.Apply(ctx); err != nil {
			return err
		}
	case trxs.TRX_PROPOSAL, trxs.TRX_VOTING:
		h := txe.handlers[2]
		if err := h.Validate(ctx); err != nil {
			return err
		} else if err := h.Apply(ctx); err != nil {
			return err
		}
	default:
		return xerrors.ErrInvalidTrxType
	}

	//for _, h := range txe.handlers {
	//	if err := h.Validate(ctx); err != nil {
	//		return err
	//	} else if err := h.Apply(ctx); err != nil {
	//		return err
	//	}
	//}
	return nil
}
