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

const (
	ErrCodeQuery uint32 = 1000 + iota
	ErrCodeInvalidQueryPath
	ErrCodeInvalidQueryParams
	ErrCodeNotFoundResult
	ErrLast
)

var (
	ErrInitChain  = Wrap(ErrCodeGeneric, "InitChain failed", nil)
	ErrCheckTx    = Wrap(ErrCodeCheckTx, "CheckTx failed", nil)
	ErrBeginBlock = Wrap(ErrCodeBeginBlock, "BeginBlock failed", nil)
	ErrDeliverTx  = Wrap(ErrCodeDeliverTx, "DeliverTx failed", nil)
	ErrEndBlock   = Wrap(ErrCodeEndBlock, "EndBlock failed", nil)
	ErrCommit     = Wrap(ErrCodeCommit, "Commit failed", nil)
	ErrQuery      = Wrap(ErrCodeQuery, "query failed", nil)

	ErrNotFoundAccount         = Wrap(ErrCodeNotFoundAccount, "not found account", nil)
	ErrInvalidAccountType      = Wrap(ErrCodeInvalidAccountType, "invalid account type", nil)
	ErrInvalidTrx              = Wrap(ErrCodeInvalidTrx, "invalid transaction", nil)
	ErrNegFee                  = ErrInvalidTrx.Wrap(New("negative fee"))
	ErrInsufficientFee         = ErrInvalidTrx.Wrap(New("not enough fee"))
	ErrInvalidNonce            = ErrInvalidTrx.Wrap(New("invalid nonce"))
	ErrNegAmount               = ErrInvalidTrx.Wrap(New("negative amount"))
	ErrInsufficientFund        = ErrInvalidTrx.Wrap(New("insufficient fund"))
	ErrInvalidTrxType          = ErrInvalidTrx.Wrap(New("wrong transaction type"))
	ErrInvalidTrxPayloadType   = ErrInvalidTrx.Wrap(New("wrong transaction payload type"))
	ErrInvalidTrxPayloadParams = ErrInvalidTrx.Wrap(New("invalid params of transaction payload"))
	ErrInvalidTrxSig           = ErrInvalidTrx.Wrap(New("invalid signature"))
	ErrTooManyPower            = ErrInvalidTrx.Wrap(New("too many power"))
	ErrNotFoundTx              = Wrap(ErrCodeNotFoundTx, "not found tx", nil)
	ErrNotFoundDelegatee       = Wrap(ErrCodeNotFoundDelegatee, "not found delegatee", nil)
	ErrNotFoundStake           = Wrap(ErrCodeNotFoundStake, "not found stake", nil)
	ErrNotFoundProposal        = Wrap(ErrCodeNotFoundProposal, "not found proposal", nil)

	ErrInvalidQueryPath   = Wrap(ErrCodeInvalidQueryPath, "invalid query path", nil)
	ErrInvalidQueryParams = Wrap(ErrCodeInvalidQueryParams, "invalid query parameters", nil)

	ErrNotFoundResult = Wrap(ErrCodeNotFoundResult, "not found result", nil)

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
	Wrap(error) XError
}

type xerror struct {
	code  uint32
	msg   string
	cause error
}

func New(msg string) XError {
	return &xerror{
		code: ErrCodeGeneric,
		msg:  msg,
	}
}

func From(err error) XError {
	return New(err.Error())
}

func Wrap(code uint32, msg string, causer error) XError {
	return &xerror{
		code:  code,
		msg:   msg,
		cause: causer,
	}
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
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

//func (e *xerror) Unwrap() error {
//	return e.Cause()
//}
//
//func (e *xerror) With(err error) XError {
//	return &xerror{
//		code:  e.code,
//		msg:   e.msg,
//		cause: err,
//	}
//}
//

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
