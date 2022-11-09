package gov

import (
	"encoding/hex"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"math/big"
	"strings"
)

type stakerHelper struct {
	powers map[string]int64
}

func (s *stakerHelper) GetTotalAmount() *big.Int {
	return govRules.PowerToAmount(s.GetTotalPower())
}

func (s *stakerHelper) GetTotalPower() int64 {
	sum := int64(0)
	for _, v := range s.powers {
		sum += v
	}
	return sum
}

func (s *stakerHelper) GetTotalPowerOf(addr types.Address) int64 {
	return s.powers[addr.String()]
}

func (s *stakerHelper) IsValidator(addr types.Address) bool {
	addrStr := addr.String()
	for k, _ := range s.powers {
		if strings.Compare(k, addrStr) == 0 {
			return true
		}
	}
	return false
}

func (s *stakerHelper) PickAddress(i int) types.Address {
	var addrs []string
	for k, _ := range s.powers {
		addrs = append(addrs, k)
	}
	ret, _ := hex.DecodeString(addrs[i%len(s.powers)])
	return ret
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
	stakerCtrler = &stakerHelper{
		powers: map[string]int64{
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
			libs.RandAddress().String(): 10,
		},
	}
	accountCtrler    = &accountHelper{}
	accountMap       = make(map[types.AcctKey]types.IAccount)
	proposalGovRules = &GovRules{
		Version:            1,
		MaxValidatorCnt:    11,
		AmountPerPower:     big.NewInt(2_000000000_000000000),
		RewardPerPower:     big.NewInt(2_000000000),
		LazyRewardBlocks:   20,
		LazyApplyingBlocks: 20,
	}
)

func makeProposalTrxCtx(from, to types.Address, gas *big.Int) *trxs.TrxContext {
	bz, _ := proposalGovRules.Encode()
	tx := client.NewTrxGovRulesProposal(
		from, to, gas, 1, "test govrules proposal", bz)

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

func makeVotingTrxCtx(proposalTxHash types.HexBytes, choice int32) *trxs.TrxContext {
	tx := client.NewTrxVoting(
		libs.RandAddress(), libs.ZeroBytes(types.AddrSize), big.NewInt(10), 1, proposalTxHash, choice)

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
