package types

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	rtypes "github.com/rigochain/rigo-go/types"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
	"math"
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
	GasUsed      uint64
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
	}

	txctx := &TrxContext{
		Tx:        tx,
		TxHash:    types.Tx(txbz).Hash(),
		Height:    height,
		BlockTime: btime,
		Exec:      exec,
		GasUsed:   0,
	}
	for _, fn := range cbfns {
		if err := fn(txctx); err != nil {
			return nil, err
		}
	}

	if xerr := txctx.Validate(); xerr != nil {
		return nil, xerr
	}

	return txctx, nil
}

func (txctx *TrxContext) Validate() xerrors.XError {
	tx := txctx.Tx

	if len(tx.From) != rtypes.AddrSize {
		return xerrors.ErrInvalidAddress
	}
	if len(tx.To) != rtypes.AddrSize {
		return xerrors.ErrInvalidAddress
	}
	if tx.Amount.Sign() < 0 {
		return xerrors.ErrInvalidAmount
	}
	if tx.Gas < 0 || tx.Gas > math.MaxInt64 {
		return xerrors.ErrInvalidGas
	}
	if tx.GasPrice.Sign() < 0 || tx.GasPrice.Cmp(txctx.GovHandler.GasPrice()) < 0 {
		return xerrors.ErrInvalidGasPrice
	}

	// check signature

	if txctx.Exec {
		sig := tx.Sig
		tx.Sig = nil
		_txbz, err := rlp.EncodeToBytes(tx)
		tx.Sig = sig
		if err != nil {
			return xerrors.From(err)
		}
		fromAddr, pubBytes, xerr := crypto.Sig2Addr(_txbz, sig)
		if xerr != nil {
			return xerr
		}
		if bytes.Compare(fromAddr, tx.From) != 0 {
			return xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, fromAddr))
		}
		txctx.SenderPubKey = pubBytes
	}

	txctx.Sender = txctx.AcctHandler.FindAccount(tx.From, txctx.Exec)
	if txctx.Sender == nil {
		return xerrors.ErrNotFoundAccount
	}
	if !rtypes.IsZeroAddress(tx.To) {
		txctx.Receiver = txctx.AcctHandler.FindOrNewAccount(tx.To, txctx.Exec)
	}

	needAmt := new(uint256.Int).Mul(tx.GasPrice, uint256.NewInt(uint64(tx.Gas)))
	_ = needAmt.Add(needAmt, tx.Amount)
	if xerr := txctx.Sender.CheckBalance(needAmt); xerr != nil {
		return xerr
	}
	if xerr := txctx.Sender.CheckNonce(tx.Nonce); xerr != nil {
		return xerr.Wrap(fmt.Errorf("invalid nonce - ledger: %v, tx:%v, address: %v, txhash: %X", txctx.Sender.GetNonce(), tx.Nonce, txctx.Sender.Address, txctx.TxHash))
	}

	return nil
}
