package xerrors

import (
	"errors"
	"fmt"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

const (
	ErrCodeSuccess uint32 = abcitypes.CodeTypeOK + iota
	ErrCodeOrdinary
	ErrCodeInitChain
	ErrCodeCheckTx
	ErrCodeBeginBlock
	ErrCodeDeliverTx
	ErrCodeEndBlock
	ErrCodeCommit
	ErrCodeNotFoundAccount
	ErrCodeInvalidAccountType
	ErrCodeInvalidTrx
	ErrCodeNotFoundTx
	ErrCodeNotFoundDelegatee
	ErrCodeNotFoundStake
	ErrCodeNotFoundProposal
)

const (
	ErrCodeQuery uint32 = 1000 + iota
	ErrCodeInvalidQueryPath
	ErrCodeInvalidQueryParams
	ErrCodeNotFoundResult
	ErrLast
)

var (
	ErrCommon     = New(ErrCodeOrdinary, "rigo error", nil)
	ErrInitChain  = New(ErrCodeInitChain, "InitChain failed", nil)
	ErrCheckTx    = New(ErrCodeCheckTx, "CheckTx failed", nil)
	ErrBeginBlock = New(ErrCodeBeginBlock, "BeginBlock failed", nil)
	ErrDeliverTx  = New(ErrCodeDeliverTx, "DeliverTx failed", nil)
	ErrEndBlock   = New(ErrCodeEndBlock, "EndBlock failed", nil)
	ErrCommit     = New(ErrCodeCommit, "Commit failed", nil)
	ErrQuery      = New(ErrCodeQuery, "query failed", nil)

	ErrNotFoundAccount         = New(ErrCodeNotFoundAccount, "not found account", nil)
	ErrInvalidAccountType      = New(ErrCodeInvalidAccountType, "invalid account type", nil)
	ErrInvalidTrx              = New(ErrCodeInvalidTrx, "invalid transaction", nil)
	ErrNegFee                  = ErrInvalidTrx.Wrap(errors.New("negative fee"))
	ErrInsufficientFee         = ErrInvalidTrx.Wrap(errors.New("not enough fee"))
	ErrInvalidAddress          = ErrInvalidTrx.Wrap(errors.New("invalid address"))
	ErrInvalidNonce            = ErrInvalidTrx.Wrap(errors.New("invalid nonce"))
	ErrNegAmount               = ErrInvalidTrx.Wrap(errors.New("negative amount"))
	ErrInsufficientFund        = ErrInvalidTrx.Wrap(errors.New("insufficient fund"))
	ErrInvalidTrxType          = ErrInvalidTrx.Wrap(errors.New("wrong transaction type"))
	ErrInvalidTrxPayloadType   = ErrInvalidTrx.Wrap(errors.New("wrong transaction payload type"))
	ErrInvalidTrxPayloadParams = ErrInvalidTrx.Wrap(errors.New("invalid params of transaction payload"))
	ErrInvalidTrxSig           = ErrInvalidTrx.Wrap(errors.New("invalid signature"))
	ErrNotFoundTx              = New(ErrCodeNotFoundTx, "not found tx", nil)
	ErrNotFoundDelegatee       = New(ErrCodeNotFoundDelegatee, "not found delegatee", nil)
	ErrNotFoundStake           = New(ErrCodeNotFoundStake, "not found stake", nil)
	ErrNotFoundProposal        = New(ErrCodeNotFoundProposal, "not found proposal", nil)

	ErrInvalidQueryPath   = New(ErrCodeInvalidQueryPath, "invalid query path", nil)
	ErrInvalidQueryParams = New(ErrCodeInvalidQueryParams, "invalid query parameters", nil)

	ErrNotFoundResult = New(ErrCodeNotFoundResult, "not found result", nil)

	// new style errors
	ErrUnknownTrxType        = NewOrdinary("unknown transaction type")
	ErrUnknownTrxPayloadType = NewOrdinary("unknown transaction payload type")
	ErrNoRight               = NewOrdinary("no right")
	ErrNotVotingPeriod       = NewOrdinary("not voting period")
	ErrDuplicatedKey         = NewOrdinary("already existed key")
)

type XError interface {
	Code() uint32
	Error() string
	Cause() error
	Wrap(error) XError
	Wrapf(string, ...any) XError
}

type xerror struct {
	code  uint32
	msg   string
	cause error
}

func New(code uint32, msg string, err error) XError {
	return &xerror{
		code:  code,
		msg:   msg,
		cause: err,
	}
}

func NewOrdinary(msg string) XError {
	return &xerror{
		code: ErrCodeOrdinary,
		msg:  msg,
	}
}

func From(err error) XError {
	return NewOrdinary(err.Error())
}

func Wrap(err error, msg string) XError {
	return &xerror{
		code:  ErrCodeOrdinary,
		msg:   msg,
		cause: err,
	}
}

func Cause(err error) error {
	for err != nil {
		cause, ok := err.(XError)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func (e *xerror) Code() uint32 {
	type hascode interface {
		Cause() error
	}

	return e.code
}

func (e *xerror) Error() string {
	msg := e.msg

	if e.cause != nil {
		msg += "\n\t" + e.cause.Error()
	}

	return msg

}

func (e *xerror) Cause() error {
	return e.cause
}

func (e *xerror) Wrap(err error) XError {
	w, ok := e.cause.(XError)
	if w != nil && ok {
		err = w.Wrap(err)
	}

	return &xerror{
		code:  e.code,
		msg:   e.msg,
		cause: err,
	}
}

func (e *xerror) Wrapf(format string, args ...any) XError {
	return e.Wrap(fmt.Errorf(format, args))
}
