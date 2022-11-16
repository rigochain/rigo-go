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

var (
	stakerHandlerHelper = &stakerHandler{
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
	accountHandlerHelper = &accountHandler{}
	accountMap           = make(map[types.AcctKey]types.IAccount)
	govRuleProposal      = &GovRule{
		Version:            1,
		MaxValidatorCnt:    11,
		AmountPerPower:     big.NewInt(2_000000000_000000000),
		RewardPerPower:     big.NewInt(2_000000000),
		LazyRewardBlocks:   20,
		LazyApplyingBlocks: 20,
	}
)

type stakerHandler struct {
	powers map[string]int64
}

func (s *stakerHandler) GetTotalAmount() *big.Int {
	return govCtrler.PowerToAmount(s.GetTotalPower())
}

func (s *stakerHandler) GetTotalPower() int64 {
	sum := int64(0)
	for _, v := range s.powers {
		sum += v
	}
	return sum
}

func (s *stakerHandler) GetTotalPowerOf(addr types.Address) int64 {
	return s.powers[addr.String()]
}

func (s *stakerHandler) IsValidator(addr types.Address) bool {
	addrStr := addr.String()
	for k, _ := range s.powers {
		if strings.Compare(k, addrStr) == 0 {
			return true
		}
	}
	return false
}

func (s *stakerHandler) PickAddress(i int) types.Address {
	var addrs []string
	for k, _ := range s.powers {
		addrs = append(addrs, k)
	}
	ret, _ := hex.DecodeString(addrs[i%len(s.powers)])
	return ret
}

type accountHandler struct{}

func (a *accountHandler) FindAccount(addr types.Address, exec bool) types.IAccount {
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

func makeProposalTrxCtx(from, to types.Address, gas *big.Int, bzProposal []byte) *trxs.TrxContext {
	tx := client.NewTrxGovRuleProposal(
		from, to, gas, 1, "test govrule proposal", bzProposal)

	txbz, _ := tx.Encode()
	txctx, _ := trxs.NewTrxContextEx(txbz, 10, false, func(_txctx *trxs.TrxContext) error {
		_tx := _txctx.Tx
		// find sender account
		acct := accountHandlerHelper.FindAccount(_tx.From, _txctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount
		}
		_txctx.Sender = acct
		_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
		_txctx.GovRuleHandler = govCtrler
		_txctx.StakeHandler = stakerHandlerHelper
		return nil
	})

	return txctx
}

func makeVotingTrxCtx(from types.Address, proposalTxHash types.HexBytes, choice int32) *trxs.TrxContext {
	tx := client.NewTrxVoting(
		from, libs.ZeroBytes(types.AddrSize), big.NewInt(10), 1, proposalTxHash, choice)

	txbz, _ := tx.Encode()
	txctx, _ := trxs.NewTrxContextEx(txbz, 10, false, func(_txctx *trxs.TrxContext) error {
		_tx := _txctx.Tx
		// find sender account
		acct := accountHandlerHelper.FindAccount(_tx.From, _txctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount
		}
		_txctx.Sender = acct
		_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
		_txctx.GovRuleHandler = govCtrler
		_txctx.StakeHandler = stakerHandlerHelper
		return nil
	})

	return txctx
}
