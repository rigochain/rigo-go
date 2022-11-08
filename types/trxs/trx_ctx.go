package trxs

import (
	"bytes"
	"fmt"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
)

type TrxContext struct {
	Tx     *Trx
	TxHash types.HexBytes

	Height   int64
	Proposer types.Address
	Exec     bool

	SenderPubKey []byte
	Sender       types.IAccount
	Receiver     types.IAccount
	NeedAmt      *big.Int
	GasUsed      *big.Int

	GovRules    types.IGovRules
	StakeCtrler types.IStakeCtrler
	Error       xerrors.XError
}

//func NewTrxContext(txbz []byte, height int64, exec bool, acctFinder types.IAccountFinder, govRules types.IGovRules) (*TrxContext, error) {
//	tx := &Trx{}
//	if xerr := tx.Decode(txbz); xerr != nil {
//		return nil, xerr
//	} else if tx.Amount.Sign() <= 0 {
//		return nil, xerrors.ErrNegAmount
//	} else if tx.Gas.Sign() < 0 {
//		return nil, xerrors.ErrNegGas
//	}
//
//	// find sender account
//	acct := acctFinder.FindAccount(tx.From, exec)
//	if acct == nil {
//		return nil, xerrors.ErrNotFoundAccount
//	}
//
//	// check sender account nonce
//	if xerr := acct.CheckNonce(tx.Nonce); xerr != nil {
//		return nil, xerr
//	}
//	acct.AddNonce()
//
//	// check sender account balance
//	needFund := new(big.Int).Add(tx.Amount, tx.Gas)
//	if xerr := acct.CheckBalance(needFund); xerr != nil {
//		return nil, xerr
//	}
//
//	var pubBytes types.HexBytes
//	if exec {
//		sig := tx.Sig
//		tx.Sig = nil
//		if _txbz, err := tx.Encode(); err != nil {
//			return nil, xerrors.ErrInvalidTrx
//		} else if _addr, _pub, err := crypto.Sig2Addr(_txbz, sig); err != nil {
//			return nil, xerrors.NewFrom(err)
//		} else if bytes.Compare(_addr, tx.From) != 0 {
//			return nil, xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, _addr))
//		} else {
//			pubBytes = _pub
//		}
//	}
//
//	return &TrxContext{
//		Tx:           tx,
//		TxHash:       crypto.DefaultHash(txbz),
//		Height:       height,
//		Exec:         exec,
//		Sender:       acct,
//		SenderPubKey: pubBytes,
//		NeedAmt:      needFund,
//		GasUsed:      big.NewInt(0),
//		//AccountFinder: acctFinder,
//		GovRules: govRules,
//	}, nil
//}

type NewTrxContextCb func(*TrxContext) error

func NewTrxContextEx(txbz []byte, height int64, exec bool, cbfns ...NewTrxContextCb) (*TrxContext, error) {
	tx := &Trx{}
	if xerr := tx.Decode(txbz); xerr != nil {
		return nil, xerr
	} else if tx.Amount.Sign() < 0 {
		return nil, xerrors.ErrNegAmount
	} else if tx.Gas.Sign() < 0 {
		return nil, xerrors.ErrNegGas
	}

	var pubBytes types.HexBytes
	if exec {
		sig := tx.Sig
		tx.Sig = nil
		if _txbz, err := tx.Encode(); err != nil {
			return nil, xerrors.ErrInvalidTrx
		} else if _addr, _pub, err := crypto.Sig2Addr(_txbz, sig); err != nil {
			return nil, xerrors.NewFrom(err)
		} else if bytes.Compare(_addr, tx.From) != 0 {
			return nil, xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, _addr))
		} else {
			pubBytes = _pub
		}
	}

	txctx := &TrxContext{
		Tx:           tx,
		TxHash:       crypto.DefaultHash(txbz),
		Height:       height,
		Exec:         exec,
		SenderPubKey: pubBytes,
		GasUsed:      big.NewInt(0),
	}

	for _, fn := range cbfns {
		if err := fn(txctx); err != nil {
			return nil, err
		}
	}

	return txctx, nil
}
