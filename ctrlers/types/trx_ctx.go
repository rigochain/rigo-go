package types

import (
	bytes2 "github.com/kysee/arcanus/types/bytes"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/types"
	"math/big"
)

type TrxContext struct {
	Height int64
	TxHash bytes2.HexBytes
	Tx     *Trx
	TxIdx  int
	Exec   bool

	SenderPubKey []byte
	Sender       *Account
	Receiver     *Account
	NeedAmt      *big.Int
	GasUsed      *big.Int

	GovHandler   ITrxHandler
	AcctHandler  ITrxHandler
	StakeHandler ITrxHandler

	GovHelper   IGovHelper
	StakeHelper IStakeHelper

	Callback func(*TrxContext, xerrors.XError)
}

type ITrxHandler interface {
	ValidateTrx(*TrxContext) xerrors.XError
	ExecuteTrx(*TrxContext) xerrors.XError
}

type NewTrxContextCb func(*TrxContext) xerrors.XError

func NewTrxContext(txbz []byte, height int64, exec bool, cbfns ...NewTrxContextCb) (*TrxContext, xerrors.XError) {
	tx := &Trx{}
	if xerr := tx.Decode(txbz); xerr != nil {
		return nil, xerr
	} else if tx.Amount.Sign() < 0 {
		return nil, xerrors.ErrNegAmount
	} else if tx.Gas.Sign() < 0 {
		return nil, xerrors.ErrNegFee
	}

	txctx := &TrxContext{
		Tx:      tx,
		TxHash:  types.Tx(txbz).Hash(),
		Height:  height,
		Exec:    exec,
		GasUsed: big.NewInt(0),
	}

	for _, fn := range cbfns {
		if err := fn(txctx); err != nil {
			return nil, err
		}
	}

	return txctx, nil
}
