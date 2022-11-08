package gov

import (
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
)

type stakerHelper struct{}

func (s *stakerHelper) GetTotalAmount() *big.Int {
	return big.NewInt(0)
}

func (s *stakerHelper) GetTotalPower() int64 {
	return int64(1000)
}

func (s *stakerHelper) GetTotalPowerOf(addr types.Address) int64 {
	return int64(10)
}

type accountHelper struct{}

func (a *accountHelper) FindAccount(addr types.Address, exec bool) types.IAccount {
	acctKey := types.ToAcctKey(addr)
	if acct, ok := accountMap[acctKey]; ok {
		return acct
	} else {
		acct = account.NewAccount(addr)
		acct.AddBalance(big.NewInt(100000))
		accountMap[acctKey] = acct
		return acct
	}
}

var (
	stakerCtrler  = &stakerHelper{}
	accountCtrler = &accountHelper{}
	accountMap    = make(map[types.AcctKey]types.IAccount)
)

func makeTrxCtx() *trxs.TrxContext {
	rules := DefaultGovRules()
	bz, _ := rules.Encode()
	tx := client.NewTrxGovRulesProposal(
		libs.RandAddress(), libs.RandAddress(), big.NewInt(10), 1, "test govrules proposal", bz)

	txbz, _ := tx.Encode()
	txctx, _ := trxs.NewTrxContextEx(txbz, 10, false, func(_txctx *trxs.TrxContext) error {
		_tx := _txctx.Tx
		// find sender account
		acct := accountCtrler.FindAccount(_tx.From, _txctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount
		}
		_txctx.Sender = acct

		// check sender account nonce
		if xerr := acct.CheckNonce(_tx.Nonce); xerr != nil {
			return xerr
		}
		acct.AddNonce()

		// check sender account balance
		needFund := new(big.Int).Add(_tx.Amount, _tx.Gas)
		if xerr := acct.CheckBalance(needFund); xerr != nil {
			return xerr
		}
		_txctx.NeedAmt = needFund

		_txctx.GovRules = DefaultGovRules()
		_txctx.StakeCtrler = stakerCtrler

		return nil
	})

	return txctx
}
