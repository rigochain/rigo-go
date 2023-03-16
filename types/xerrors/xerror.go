package xerrors

import (
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
	ErrCodeNotFoundDelegatee
	ErrCodeNotFoundStake
	ErrCodeNotFoundProposal
)

type xerr2 error

const (
	ErrCodeQuery uint32 = 1000 + iota
	ErrCodeInvalidQueryPath
	ErrCodeInvalidQueryParams
	ErrCodeNotFoundResult
	ErrLast
)

var (
	ErrCheckTx    = NewWith(ErrCodeCheckTx, "CheckTx failed")
	ErrBeginBlock = NewWith(ErrCodeBeginBlock, "BeginBlock failed")
	ErrDeliverTx  = NewWith(ErrCodeDeliverTx, "DeliverTx failed")
	ErrEndBlock   = NewWith(ErrCodeEndBlock, "EndBlock failed")
	ErrCommit     = NewWith(ErrCodeCommit, "Commit failed")
	ErrQuery      = NewWith(ErrCodeQuery, "query failed")

	ErrNotFoundAccount         = NewWith(ErrCodeNotFoundAccount, "not found account")
	ErrInvalidAccountType      = NewWith(ErrCodeInvalidAccountType, "invalid account type")
	ErrInvalidTrx              = NewWith(ErrCodeInvalidTrx, "invalid transaction")
	ErrNegFee                  = ErrInvalidTrx.Wrap(New("negative fee"))
	ErrInsufficientFee         = ErrInvalidTrx.Wrap(New("not enough fee"))
	ErrInvalidNonce            = ErrInvalidTrx.Wrap(New("invalid nonce"))
	ErrNegAmount               = ErrInvalidTrx.Wrap(New("negative amount"))
	ErrInsufficientFund        = ErrInvalidTrx.Wrap(New("insufficient fund"))
	ErrInvalidTrxType          = ErrInvalidTrx.Wrap(New("wrong transaction type"))
	ErrInvalidTrxPayloadType   = ErrInvalidTrx.Wrap(New("wrong transaction payload type"))
	ErrInvalidTrxPayloadParams = ErrInvalidTrx.With(New("invalid params of transaction payload"))
	ErrInvalidTrxSig           = ErrInvalidTrx.With(New("invalid signature"))
	ErrTooManyPower            = ErrInvalidTrx.With(New("too many power"))
	ErrNotFoundTx              = NewWith(ErrCodeNotFoundTx, "not found tx")
	ErrNotFoundDelegatee       = NewWith(ErrCodeNotFoundDelegatee, "not found delegatee")
	ErrNotFoundStake           = NewWith(ErrCodeNotFoundStake, "not found stake")
	ErrNotFoundProposal        = NewWith(ErrCodeNotFoundProposal, "not found proposal")

	ErrInvalidQueryPath   = NewWith(ErrCodeInvalidQueryPath, "invalid query path")
	ErrInvalidQueryParams = NewWith(ErrCodeInvalidQueryParams, "invalid query parameters")

	ErrNotFoundResult = NewWith(ErrCodeNotFoundResult, "not found result")

	// new style errors
	ErrUnknownTrxType        = New("unknown transaction type")
	ErrUnknownTrxPayloadType = New("unknown transaction payload type")
	ErrNoRight               = New("no right")
	ErrNotVotingPeriod       = New("not voting period")
	ErrDuplicatedKey         = New("already existed key")
)

type XError interface {
	Code() uint32
	Error() string
	Cause() error
	With(error) XError
	Wrap(error) XError
	Unwrap() error
}

type xerror struct {
	code  uint32
	msg   string
	cause error
}

func New(m string) XError {
	return &xerror{
		code: ErrCodeGeneric,
		msg:  m,
	}
}

func NewWith(code uint32, msg string) XError {
	return &xerror{
		code: code,
		msg:  msg,
	}
}

func NewFrom(err error) XError {
	if err != nil {
		return &xerror{
			code: ErrCodeGeneric,
			msg:  err.Error(),
		}
	}
	return nil
}

func (e *xerror) Code() uint32 {
	return e.code
}

func (e *xerror) Error() string {
	type causer interface {
		Cause() error
	}

	msg := e.msg

	if e.cause != nil {
		msg += " << " + e.cause.Error()
	}

	return msg

}

func (e *xerror) Cause() error {
	return e.cause
}

func (e *xerror) Unwrap() error {
	return e.Cause()
}

func (e *xerror) With(err error) XError {
	return &xerror{
		code:  e.code,
		msg:   e.msg,
		cause: err,
	}
}

func (e *xerror) Wrap(err error) XError {
	type wapper interface {
		Wrap(error) XError
	}

	w, ok := e.cause.(wapper)
	if w != nil && ok {
		err = w.Wrap(err)
	}

	return &xerror{
		code:  e.code,
		msg:   e.msg,
		cause: err,
	}
}
