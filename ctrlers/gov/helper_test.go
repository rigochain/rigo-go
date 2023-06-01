package gov

import (
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"time"
)

type stakeHandlerMock struct {
	valCnt     int
	delegatees []*stake.Delegatee
}

func (s *stakeHandlerMock) Validators() ([]*abcitypes.Validator, int64) {
	totalPower := int64(0)
	vals := make([]*abcitypes.Validator, s.valCnt)
	for i := 0; i < s.valCnt; i++ {
		vals[i] = &abcitypes.Validator{
			Address: s.delegatees[i].Addr,
			Power:   s.delegatees[i].TotalPower,
		}
		totalPower += s.delegatees[i].TotalPower
	}
	return vals, totalPower
}

func (s *stakeHandlerMock) IsValidator(addr types.Address) bool {
	for i := 0; i < s.valCnt; i++ {
		if bytes.Compare(addr, s.delegatees[i].Addr) == 0 {
			return true
		}
	}
	return false
}

func (s *stakeHandlerMock) GetTotalAmount() *uint256.Int {
	return ctrlertypes.PowerToAmount(s.GetTotalPower())
}

func (s *stakeHandlerMock) GetTotalPower() int64 {
	sum := int64(0)
	for _, v := range s.delegatees {
		sum += v.TotalPower
	}
	return sum
}

func (s *stakeHandlerMock) GetTotalPowerOf(addr types.Address) int64 {
	for _, v := range s.delegatees {
		if bytes.Compare(addr, v.Addr) == 0 {
			return v.TotalPower
		}
	}
	return int64(0)
}

func (s *stakeHandlerMock) PowerOf(addr types.Address) int64 {
	return s.GetTotalPowerOf(addr)
}

func (s *stakeHandlerMock) PickAddress(i int) types.Address {
	return s.delegatees[i].Addr
}

var _ ctrlertypes.IStakeHandler = (*stakeHandlerMock)(nil)

type acctHelperMock struct {
	acctMap map[ctrlertypes.AcctKey]*ctrlertypes.Account
}

func (a *acctHelperMock) FindAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	acctKey := ctrlertypes.ToAcctKey(addr)
	if acct, ok := a.acctMap[acctKey]; ok {
		return acct
	} else {
		acct = ctrlertypes.NewAccount(addr)
		acct.AddBalance(uint256.NewInt(100000))
		a.acctMap[acctKey] = acct
		return acct
	}
}

func makeTrxCtx(tx *ctrlertypes.Trx, height int64, exec bool) *ctrlertypes.TrxContext {
	txbz, _ := tx.Encode()
	txctx, _ := ctrlertypes.NewTrxContext(txbz, height, time.Now().Unix(), exec, func(_txctx *ctrlertypes.TrxContext) xerrors.XError {
		_tx := _txctx.Tx
		// find sender account
		acct := acctHelper.FindAccount(_tx.From, _txctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount
		}
		_txctx.Sender = acct
		_txctx.NeedAmt = new(uint256.Int).Add(_tx.Amount, _tx.Gas)
		_txctx.GovHandler = govCtrler
		_txctx.StakeHandler = stakeHelper
		return nil
	})

	return txctx
}

func runCase(c *Case) xerrors.XError {
	return runTrx(c.txctx)
}

func runTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	if xerr := govCtrler.ValidateTrx(ctx); xerr != nil {
		return xerr
	}
	if xerr := govCtrler.ExecuteTrx(ctx); xerr != nil {
		return xerr
	}
	return nil
}
