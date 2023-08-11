package types

import (
	"github.com/holiman/uint256"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
)

type TrxContext struct {
	Height    int64
	BlockTime int64
	TxHash    bytes2.HexBytes
	Tx        *Trx
	TxIdx     int
	Exec      bool

	SenderPubKey []byte
	Sender       *Account
	Receiver     *Account
	NeedAmt      *uint256.Int
	GasUsed      *uint256.Int
	RetData      []byte
	Events       []abcitypes.Event

	TrxGovHandler   ITrxHandler
	TrxAcctHandler  ITrxHandler
	TrxStakeHandler ITrxHandler
	TrxEVMHandler   ITrxHandler

	GovHandler   IGovHandler
	AcctHandler  IAccountHandler
	StakeHandler IStakeHandler

	Callback func(*TrxContext, xerrors.XError)
}

type ITrxHandler interface {
	ValidateTrx(*TrxContext) xerrors.XError
	ExecuteTrx(*TrxContext) xerrors.XError
}

type NewTrxContextCb func(*TrxContext) xerrors.XError

func NewTrxContext(txbz []byte, height, btime int64, exec bool, cbfns ...NewTrxContextCb) (*TrxContext, xerrors.XError) {
	tx := &Trx{}
	if xerr := tx.Decode(txbz); xerr != nil {
		return nil, xerr
	} else if tx.Amount.Sign() < 0 {
		return nil, xerrors.ErrNegAmount
	} else if tx.Gas.Sign() < 0 {
		return nil, xerrors.ErrNegFee
	}

	txctx := &TrxContext{
		Tx:        tx,
		TxHash:    types.Tx(txbz).Hash(),
		Height:    height,
		BlockTime: btime,
		Exec:      exec,
		NeedAmt:   new(uint256.Int).Add(tx.Amount, tx.Gas),
		GasUsed:   uint256.NewInt(0),
	}

	for _, fn := range cbfns {
		if err := fn(txctx); err != nil {
			return nil, err
		}
	}

	return txctx, nil
}
