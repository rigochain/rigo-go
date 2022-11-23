package account

import (
	"github.com/cosmos/iavl"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
	"math/big"
	"sort"
	"sync"
)

type AccountCtrler struct {
	acctDB   db.DB
	acctTree *iavl.MutableTree

	simuAccounts map[account.AcctKey]*account.Account
	execAccounts map[account.AcctKey]*account.Account

	logger log.Logger
	mtx    sync.RWMutex
}

func NewAccountCtrler(dbDir string, logger log.Logger) (*AccountCtrler, error) {
	acctDB, err := db.NewDB("account", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}
	acctTree, err := iavl.NewMutableTree(acctDB, 128)
	if err != nil {
		return nil, err
	}
	if _, err := acctTree.Load(); err != nil {
		return nil, err
	}

	ret := &AccountCtrler{
		acctDB:       acctDB,
		acctTree:     acctTree,
		simuAccounts: make(map[account.AcctKey]*account.Account),
		execAccounts: make(map[account.AcctKey]*account.Account),
		logger:       logger,
	}
	return ret, nil
}

func (ctrler *AccountCtrler) PutAccount(acct *account.Account, exec bool) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.putAccount(acct, exec)
}

func (ctrler *AccountCtrler) putAccount(acct *account.Account, exec bool) {
	acctSet := ctrler.simuAccounts
	if exec {
		acctSet = ctrler.execAccounts
	}
	acctSet[acct.Key()] = acct
}

func (ctrler *AccountCtrler) FindAccount(addr account.Address, exec bool) *account.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findAccount(addr, exec)
}

func (ctrler *AccountCtrler) FindOrNewAccount(addr account.Address, exec bool) *account.Account {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.findOrNewAccount(addr, exec)
}

func (ctrler *AccountCtrler) findAccount(addr account.Address, exec bool) *account.Account {
	acctSet := ctrler.simuAccounts
	if exec {
		acctSet = ctrler.execAccounts
	}

	k := account.ToAcctKey(addr)
	if acct, ok := acctSet[k]; ok {
		return acct
	}

	if acct := ctrler.readAccount(addr); acct != nil {
		acctSet[k] = acct
		return acct
	}

	return nil
}

func (ctrler *AccountCtrler) findOrNewAccount(addr account.Address, exec bool) *account.Account {
	if acct := ctrler.findAccount(addr, exec); acct != nil {
		return acct
	}

	acct := account.NewAccountWithName(addr, "")
	ctrler.putAccount(acct, exec)
	return acct
}

func (ctrler *AccountCtrler) ReadAccount(addr account.Address) *account.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.readAccount(addr)
}

func (ctrler *AccountCtrler) readAccount(addr account.Address) *account.Account {
	if bz, err := ctrler.acctTree.Get(addr); err != nil {
		panic(err)
	} else if bz == nil {
		return nil
	} else if acct, err := DecodeAccount(bz); err != nil {
		panic(err)
	} else {
		return acct
	}
}

func (ctrler *AccountCtrler) Transfer(from, to *account.Account, amt *big.Int) error {
	// don't need locking,
	// because assetAccount does its own locking

	if err := from.SubBalance(amt); err != nil {
		return err
	}
	if err := to.AddBalance(amt); err != nil {
		_ = from.AddBalance(amt) // refund
		return err
	}
	return nil
}

func (ctrler *AccountCtrler) Commit() ([]byte, int64, error) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// accounts must be saved in order of their key.
	// if the order of accounts is not same,
	// each node has different iavl tree root hash.
	var acctKeys account.AcctKeyList
	for ak := range ctrler.execAccounts {
		acctKeys = append(acctKeys, ak)
	}
	sort.Sort(acctKeys)

	for _, k := range acctKeys {
		acct, ok := ctrler.execAccounts[k]
		if !ok {
			return nil, 0, xerrors.New("not found account - execAccounts may have some problem")
		}
		if v, err := EncodeAccount(acct); err != nil {
			// todo: implements recovery ??? or applying batch...
			return nil, 0, err
		} else {
			// DON'T USE the variable 'k' directly.
			// Next iteration, when the k's value is updated to next value,
			// the key of ctrler.acctTree will be updated too.
			var vk account.AcctKey
			copy(vk[:], k[:])
			ctrler.acctTree.Set(vk[:], v)
		}
	}

	ctrler.simuAccounts = ctrler.execAccounts
	ctrler.execAccounts = make(map[account.AcctKey]*account.Account)

	return ctrler.acctTree.SaveVersion()
}

func (ctrler *AccountCtrler) Close() error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.acctDB != nil {
		if err := ctrler.acctDB.Close(); err != nil {
			return nil
		}
	}

	ctrler.acctDB = nil
	ctrler.acctTree = nil
	ctrler.simuAccounts = nil
	ctrler.simuAccounts = nil
	return nil
}

func (ctrler *AccountCtrler) Validate(ctx *trxs.TrxContext) error {
	if ctx.Tx.GetType() != trxs.TRX_TRANSFER {
		return xerrors.ErrInvalidTrxType
	}
	return nil
}

func (ctrler *AccountCtrler) Execute(ctx *trxs.TrxContext) error {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	receiverAcct := ctrler.findOrNewAccount(ctx.Tx.To, ctx.Exec)
	if err := ctrler.Transfer(ctx.Sender, receiverAcct, ctx.Tx.Amount); err != nil {
		return err
	}
	return nil
}

var _ trxs.ITrxHandler = (*AccountCtrler)(nil)
var _ types.ILedgerHandler = (*AccountCtrler)(nil)
