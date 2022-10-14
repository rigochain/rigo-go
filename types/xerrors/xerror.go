package xerrors

import (
	"errors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

const (
	ErrCodeSuccess uint32 = abcitypes.CodeTypeOK + iota
	ErrCodeGeneric
	ErrCodeCheckTx
	ErrCodeBeginBlock
	ErrCodeDeliverTx
	ErrCodeEndBlock
	ErrCodeCommit
	ErrCodeNotFoundAccount
	ErrCodeInvalidAccountType
	ErrCodeInvalidTrx
	ErrCodeNotFoundTx
	ErrCodeNotFoundStaker
	ErrCodeNotFoundStake
)

type xerr2 error

const (
	ErrCodeQuery uint32 = 1000 + iota
	ErrCodeInvalidQueryCmd
	ErrCodeInvalidQueryParams
	ErrLast
)

var (
	ErrCheckTx    = NewWith(ErrCodeCheckTx, "CheckTx failed")
	ErrBeginBlock = NewWith(ErrCodeBeginBlock, "BeginBlock failed")
	ErrDeliverTx  = NewWith(ErrCodeDeliverTx, "DeliverTx failed")
	ErrEndBlock   = NewWith(ErrCodeEndBlock, "EndBlock failed")
	ErrCommit     = NewWith(ErrCodeCommit, "Commit failed")
	ErrQuery      = NewWith(ErrCodeQuery, "query failed")

	ErrNotFoundAccount       = NewWith(ErrCodeNotFoundAccount, "not found account")
	ErrInvalidAccountType    = NewWith(ErrCodeInvalidAccountType, "invalid account type")
	ErrInvalidTrx            = NewWith(ErrCodeInvalidTrx, "invalid transaction")
	ErrNegGas                = ErrInvalidTrx.Wrap(errors.New("negative gas"))
	ErrInvalidNonce          = ErrInvalidTrx.Wrap(errors.New("invalid nonce"))
	ErrNegAmount             = ErrInvalidTrx.Wrap(errors.New("negative amount"))
	ErrInsufficientFund      = ErrInvalidTrx.Wrap(errors.New("insufficient fund"))
	ErrInvalidTrxType        = ErrInvalidTrx.Wrap(errors.New("wrong transaction type"))
	ErrInvalidTrxPayloadType = ErrInvalidTrx.Wrap(errors.New("wrong transaction payload type"))
	ErrInvalidTrxSig         = ErrInvalidTrx.With(errors.New("invalid signature"))
	ErrNotFoundTx            = NewWith(ErrCodeNotFoundTx, "not found tx")
	ErrNotFoundStaker        = NewWith(ErrCodeNotFoundStaker, "not found staker")
	ErrNotFoundStake         = NewWith(ErrCodeNotFoundStake, "not found stake")

	ErrInvalidQueryCmd    = NewWith(ErrCodeInvalidQueryCmd, "invalid query command")
	ErrInvalidQueryParams = NewWith(ErrCodeInvalidQueryParams, "invalid query parameters")
)

type XError interface {
	Code() uint32
	Error() string
	Cause() error
	With(error) XError
	Wrap(error) XError
	Unwrap() error
}

type xerr struct {
	code  uint32
	msg   string
	cause error
}

func New(m string) XError {
	return &xerr{
		code: ErrCodeGeneric,
		msg:  m,
	}
}

func NewWith(code uint32, msg string) XError {
	return &xerr{
		code: code,
		msg:  msg,
	}
}

func NewFrom(err error) XError {
	return &xerr{
		code: ErrCodeGeneric,
		msg:  err.Error(),
	}
}

func (e *xerr) Code() uint32 {
	return e.code
}

func (e *xerr) Error() string {
	if e.cause != nil {
		return e.msg + "<<" + e.cause.Error()
	}
	return e.msg
}

func (e *xerr) Cause() error {
	return e.cause
}

func (e *xerr) Unwrap() error {
	return e.Cause()
}

func (e *xerr) With(err error) XError {
	return &xerr{
		code:  e.code,
		msg:   e.msg,
		cause: err,
	}
}

func (e *xerr) Wrap(err error) XError {
	return &xerr{
		code:  e.code,
		msg:   e.msg,
		cause: err,
	}
}
