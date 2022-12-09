package gov

import (
	cfg "github.com/kysee/arcanus/cmd/config"
	"github.com/kysee/arcanus/ctrlers/stake"
	ctrlertypes "github.com/kysee/arcanus/ctrlers/types"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/bytes"
	"github.com/kysee/arcanus/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/big"
	"os"
	"path/filepath"
	"sort"
)

var (
	config      = cfg.DefaultConfig()
	govCtrler   *GovCtrler
	stakeHelper *stakeHelperMock
	acctHelper  *acctHelperMock
	govRule0    = ctrlertypes.DefaultGovRule()
	govRule1    = ctrlertypes.Test1GovRule()
	govRule2    = ctrlertypes.Test2GovRule()
)

func init() {
	config.DBPath = filepath.Join(os.TempDir(), "gov-ctrler-test")
	os.RemoveAll(config.DBPath)
	os.MkdirAll(config.DBPath, 0700)

	var err error
	if govCtrler, err = NewGovCtrler(config, tmlog.NewNopLogger()); err != nil {
		panic(err)
	}
	govCtrler.GovRule = *govRule0

	stakeHelper = &stakeHelperMock{
		valCnt: 5,
		delegatees: []*stake.Delegatee{
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: bytes.RandInt63n(1000000)},
		},
	}
	sort.Sort(stake.PowerOrderDelegatees(stakeHelper.delegatees))

	acctHelper = &acctHelperMock{
		acctMap: make(map[ctrlertypes.AcctKey]*ctrlertypes.Account),
	}
}

type stakeHelperMock struct {
	valCnt     int
	delegatees []*stake.Delegatee
}

func (s *stakeHelperMock) Validators() ([]*abcitypes.Validator, int64) {
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

func (s *stakeHelperMock) IsValidator(addr types.Address) bool {
	for i := 0; i < s.valCnt; i++ {
		if bytes.Compare(addr, s.delegatees[i].Addr) == 0 {
			return true
		}
	}
	return false
}

func (s *stakeHelperMock) GetTotalAmount() *big.Int {
	return govCtrler.PowerToAmount(s.GetTotalPower())
}

func (s *stakeHelperMock) GetTotalPower() int64 {
	sum := int64(0)
	for _, v := range s.delegatees {
		sum += v.TotalPower
	}
	return sum
}

func (s *stakeHelperMock) GetTotalPowerOf(addr types.Address) int64 {
	for _, v := range s.delegatees {
		if bytes.Compare(addr, v.Addr) == 0 {
			return v.TotalPower
		}
	}
	return int64(0)
}

func (s *stakeHelperMock) PowerOf(addr types.Address) int64 {
	return s.GetTotalPowerOf(addr)
}

func (s *stakeHelperMock) PickAddress(i int) types.Address {
	return s.delegatees[i].Addr
}

var _ ctrlertypes.IStakeHelper = (*stakeHelperMock)(nil)

type acctHelperMock struct {
	acctMap map[ctrlertypes.AcctKey]*ctrlertypes.Account
}

func (a *acctHelperMock) FindAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	acctKey := ctrlertypes.ToAcctKey(addr)
	if acct, ok := a.acctMap[acctKey]; ok {
		return acct
	} else {
		acct = ctrlertypes.NewAccount(addr)
		acct.AddBalance(big.NewInt(100000))
		a.acctMap[acctKey] = acct
		return acct
	}
}

func makeTrxCtx(tx *ctrlertypes.Trx, height int64, exec bool) *ctrlertypes.TrxContext {
	txbz, _ := tx.Encode()
	txctx, _ := ctrlertypes.NewTrxContext(txbz, height, exec, func(_txctx *ctrlertypes.TrxContext) xerrors.XError {
		_tx := _txctx.Tx
		// find sender account
		acct := acctHelper.FindAccount(_tx.From, _txctx.Exec)
		if acct == nil {
			return xerrors.ErrNotFoundAccount
		}
		_txctx.Sender = acct
		_txctx.NeedAmt = new(big.Int).Add(_tx.Amount, _tx.Gas)
		_txctx.GovHelper = govCtrler
		_txctx.StakeHelper = stakeHelper
		return nil
	})

	return txctx
}
