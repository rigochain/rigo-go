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
)

type xerr2 error

const (
	ErrCodeQuery uint32 = 1000 + iota
	ErrCodeInvalidQueryCmd
	ErrCodeInvalidQueryParams
	ErrLast
)

var (
	ErrCheckTx    = WithCode(ErrCodeCheckTx, "CheckTx failed")
	ErrBeginBlock = WithCode(ErrCodeBeginBlock, "BeginBlock failed")
	ErrDeliverTx  = WithCode(ErrCodeDeliverTx, "DeliverTx failed")
	ErrEndBlock   = WithCode(ErrCodeEndBlock, "EndBlock failed")
	ErrCommit     = WithCode(ErrCodeCommit, "Commit failed")
	ErrQuery      = WithCode(ErrCodeQuery, "query failed")

	ErrNotFoundAccount       = WithCode(ErrCodeNotFoundAccount, "not found account")
	ErrInvalidAccountType    = WithCode(ErrCodeInvalidAccountType, "invalid account type")
	ErrInvalidTrx            = WithCode(ErrCodeInvalidTrx, "invalid transaction")
	ErrNegGas                = ErrInvalidTrx.Wrap(errors.New("negative gas"))
	ErrInvalidNonce          = ErrInvalidTrx.Wrap(errors.New("invalid nonce"))
	ErrNegAmount             = ErrInvalidTrx.Wrap(errors.New("negative amount"))
	ErrInsufficientFund      = ErrInvalidTrx.Wrap(errors.New("insufficient fund"))
	ErrInvalidTrxType        = ErrInvalidTrx.Wrap(errors.New("wrong transaction type"))
	ErrInvalidTrxPayloadType = ErrInvalidTrx.Wrap(errors.New("wrong transaction payload type"))
	ErrInvalidTrxSig         = ErrInvalidTrx.With(errors.New("invalid signature"))
	ErrNotFoundTx            = WithCode(ErrCodeNotFoundTx, "not found tx")
	ErrNotFoundStaker        = WithCode(ErrCodeNotFoundStaker, "not found staker")

	ErrInvalidQueryCmd    = WithCode(ErrCodeInvalidQueryCmd, "invalid query command")
	ErrInvalidQueryParams = WithCode(ErrCodeInvalidQueryParams, "invalid query parameters")
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

func WithCode(code uint32, msg string) XError {
	return &xerr{
		code: code,
		msg:  msg,
	}
}

func Wrap(err error) XError {
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
