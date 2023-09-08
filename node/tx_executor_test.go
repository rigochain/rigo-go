package node

import (
	"fmt"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type caseObj struct {
	txctx *ctrlertypes.TrxContext
	err   xerrors.XError
}

var (
	cases     []*caseObj
	govParams = ctrlertypes.DefaultGovRule()
)

func Test_commonValidation(t *testing.T) {
	for i, c := range cases {
		xerr := testCommonValidation(c.txctx)
		if c.err != nil {
			require.Equal(t, c.err.Code(), xerr.Code(), fmt.Sprintf("case #%d", i))
		} else { // success
			require.NoError(t, xerr, fmt.Sprintf("case #%d", i))
		}
	}
}

func init() {
	w0 := web3.NewWallet(nil)
	w1 := web3.NewWallet(nil)

	//
	// Small Gas
	tx := web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(99_999), govParams.GasPrice(), uint256.NewInt(0))
	_, _, _ = w0.SignTrxRLP(tx)
	txctx := makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidGas})

	//
	// Wrong GasPrice
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(100_000), govParams.GasPrice(), uint256.NewInt(0))
	_, _, _ = w0.SignTrxRLP(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidGasPrice})

	//
	// Wrong Signature - sign with proto encoding
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(100_000), govParams.GasPrice(), uint256.NewInt(0))
	_, _, _ = w0.SignTrxProto(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidTrxSig})

	//
	// Wrong Signature - no signature
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(100_000), govParams.GasPrice(), uint256.NewInt(0))
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidTrxSig})

	//
	// Wrong Signature - other's signature
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(100_000), govParams.GasPrice(), uint256.NewInt(0))
	_, _, _ = w1.SignTrxRLP(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidTrxSig})

	//
	// Invalid nonce
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 1, uint64(100_000), govParams.GasPrice(), uint256.NewInt(1000))
	_, _, _ = w0.SignTrxRLP(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInvalidNonce})

	//
	// Insufficient fund
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 0, uint64(100_000), govParams.GasPrice(), uint256.NewInt(1001))
	_, _, _ = w0.SignTrxRLP(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, xerrors.ErrInsufficientFund})

	//
	// Success
	tx = web3.NewTrxTransfer(w0.Address(), w1.Address(), 0, uint64(100_000), govParams.GasPrice(), uint256.NewInt(1000))
	_, _, _ = w0.SignTrxRLP(tx)
	txctx = makeTrxCtx(tx, 1)
	cases = append(cases, &caseObj{txctx, nil})
}

func makeTrxCtx(tx *ctrlertypes.Trx, height int64) *ctrlertypes.TrxContext {
	bz, _ := tx.Encode()
	txctx, _ := ctrlertypes.NewTrxContext(bz, height, time.Now().UnixMilli(), true, func(_txctx *ctrlertypes.TrxContext) xerrors.XError {
		_txctx.GovHandler = govParams
		_txctx.AcctHandler = &acctHandlerMock{}
		return nil
	})
	return txctx
}

func testCommonValidation(ctx *ctrlertypes.TrxContext) xerrors.XError {
	if xerr := commonValidation0(ctx); xerr != nil {
		return xerr
	}
	return commonValidation1(ctx)
}

type acctHandlerMock struct{}

func (a *acctHandlerMock) FindOrNewAccount(address types.Address, b bool) *ctrlertypes.Account {
	return a.FindAccount(address, b)
}

func (a *acctHandlerMock) FindAccount(address types.Address, b bool) *ctrlertypes.Account {
	acct := ctrlertypes.NewAccount(address)
	acct.AddBalance(govParams.MinTrxFee())
	acct.AddBalance(uint256.NewInt(1000))
	return acct
}

func (a *acctHandlerMock) Transfer(address types.Address, address2 types.Address, u *uint256.Int, b bool) xerrors.XError {
	//TODO implement me
	panic("implement me")
}

func (a *acctHandlerMock) Reward(address types.Address, u *uint256.Int, b bool) xerrors.XError {
	//TODO implement me
	panic("implement me")
}

func (a *acctHandlerMock) ImmutableAcctCtrlerAt(i int64) (ctrlertypes.IAccountHandler, xerrors.XError) {
	//TODO implement me
	panic("implement me")
}

func (a *acctHandlerMock) SetAccountCommittable(account *ctrlertypes.Account, b bool) xerrors.XError {
	//TODO implement me
	panic("implement me")
}

var _ ctrlertypes.IAccountHandler = (*acctHandlerMock)(nil)
