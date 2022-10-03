package trxs

import (
	"bytes"
	"fmt"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
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
	Error        xerrors.XError
}

func NewTrxContext(txbz []byte, height int64, exec bool, finder types.IAccountFinder) (*TrxContext, error) {
	tx := &Trx{}
	if xerr := tx.Decode(txbz); xerr != nil {
		return nil, xerr
	} else if tx.Amount.Sign() <= 0 {
		return nil, xerrors.ErrNegAmount
	} else if tx.Gas.Sign() < 0 {
		return nil, xerrors.ErrNegGas
	}

	// find sender account
	acct := finder.FindAccount(tx.From, exec)
	if acct == nil {
		return nil, xerrors.ErrNotFoundAccount
	}

	// todo: check sender account nonce
	// To test easily, do not check nonce.
	//// check sender account nonce
	//if xerr := acct.CheckNonce(tx.Nonce); xerr != nil {
	//	return nil, xerr
	//}

	// check sender account balance
	needFund := new(big.Int).Add(tx.Amount, tx.Gas)
	if xerr := acct.CheckBalance(needFund); xerr != nil {
		return nil, xerr
	}

	var pubBytes types.HexBytes
	if exec {
		sig := tx.Sig
		tx.Sig = nil
		if _txbz, err := tx.Encode(); err != nil {
			return nil, xerrors.ErrInvalidTrx
		} else if _addr, _pub, err := crypto.Sig2Addr(_txbz, sig); err != nil {
			return nil, xerrors.Wrap(err)
		} else if bytes.Compare(_addr, tx.From) != 0 {
			return nil, xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, _addr))
		} else {
			pubBytes = _pub
		}

	}

	return &TrxContext{
		Tx:           tx,
		TxHash:       tmcrypto.Sha256(txbz),
		Height:       height,
		Exec:         exec,
		Sender:       acct,
		SenderPubKey: pubBytes,
		NeedAmt:      needFund,
		GasUsed:      big.NewInt(0),
	}, nil
}
