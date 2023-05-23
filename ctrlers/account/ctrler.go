package account

import (
	"fmt"
	"github.com/holiman/uint256"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	atypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/genesis"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"sync"
)

type AcctCtrler struct {
	acctLedger ledger.IFinalityLedger[*atypes.Account]

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewAcctCtrler(config *cfg.Config, logger tmlog.Logger) (*AcctCtrler, error) {
	if execLedger, err := ledger.NewFinalityLedger[*atypes.Account]("accounts", config.DBDir(), 128, func() *atypes.Account { return &atypes.Account{} }); err != nil {
		return nil, err
	} else {
		return &AcctCtrler{
			acctLedger: execLedger,
			logger:     logger.With("module", "rigo_AcctCtrler"),
		}, nil
	}
}

func (ctrler *AcctCtrler) ImmutableAcctCtrlerAt(height int64) (atypes.IAccountHandler, xerrors.XError) {
	ledger0, xerr := ctrler.acctLedger.ImmutableLedgerAt(height, 128)
	if xerr != nil {
		return nil, xerr
	}

	return &AcctCtrler{
		acctLedger: ledger0,
		logger:     ctrler.logger,
	}, nil
}

func (ctrler *AcctCtrler) InitLedger(req interface{}) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	genAppState, ok := req.(*genesis.GenesisAppState)
	if !ok {
		return xerrors.ErrInitChain.Wrapf("wrong parameter: AcctCtrler::InitLedger requires *genesis.GenesisAppState")
	}

	for _, holder := range genAppState.AssetHolders {
		addr := append(holder.Address, nil...)
		acct := &atypes.Account{
			Address: addr,
			Balance: holder.Balance.Clone(),
		}
		if xerr := ctrler.setAccountCommittable(acct, true); xerr != nil {
			return xerr
		}
	}
	return nil
}

func (ctrler *AcctCtrler) ValidateTrx(ctx *atypes.TrxContext) xerrors.XError {
	ctx.Sender = ctrler.FindAccount(ctx.Tx.From, ctx.Exec)
	if ctx.Sender == nil {
		return xerrors.ErrNotFoundAccount
	}
	if !types.IsZeroAddress(ctx.Tx.To) {
		ctx.Receiver = ctrler.FindOrNewAccount(ctx.Tx.To, ctx.Exec)
	}

	if xerr := ctx.Sender.CheckBalance(ctx.NeedAmt); xerr != nil {
		return xerr
	}
	if xerr := ctx.Sender.CheckNonce(ctx.Tx.Nonce); xerr != nil {
		return xerr.Wrap(fmt.Errorf("invalid nonce - ledger: %v, tx:%v, address: %v, txhash: %X", ctx.Sender.GetNonce(), ctx.Tx.Nonce, ctx.Sender.Address, ctx.TxHash))
	}

	return nil
}

func (ctrler *AcctCtrler) ExecuteTrx(ctx *atypes.TrxContext) xerrors.XError {
	// Remove Lock()/Unlock() or Use RLock()/RUlock() to improve performance
	// Lock()/Unlock() make txs to be processed serially
	//ctrler.mtx.Lock()
	//defer ctrler.mtx.Unlock()

	// amount + fee
	if xerr := ctx.Sender.SubBalance(ctx.NeedAmt); xerr != nil {
		return xerr
	}

	if ctx.Tx.Type == atypes.TRX_TRANSFER {
		if xerr := ctx.Receiver.AddBalance(ctx.Tx.Amount); xerr != nil {
			return xerr
		}
	}

	// increase sender's nonce
	ctx.Sender.AddNonce()

	// set used gas
	ctx.GasUsed = ctx.Tx.Gas

	_ = ctrler.setAccountCommittable(ctx.Sender, ctx.Exec)
	if ctx.Receiver != nil {
		_ = ctrler.setAccountCommittable(ctx.Receiver, ctx.Exec)
	}

	return nil
}

func (ctrler *AcctCtrler) BeginBlock(ctx *atypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	// do nothing
	return nil, nil
}

func (ctrler *AcctCtrler) EndBlock(ctx *atypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	header := ctx.BlockInfo().Header
	if header.GetProposerAddress() != nil && ctx.GasSum().Sign() > 0 {
		// give fee to block proposer
		acct, xerr := ctrler.acctLedger.GetFinality(ledger.ToLedgerKey(header.GetProposerAddress()))
		if xerr != nil {
			return nil, xerr
		}
		xerr = acct.AddBalance(ctx.GasSum())
		if xerr != nil {
			return nil, xerr
		}
		return nil, ctrler.setAccountCommittable(acct, true)
	}
	return nil, nil
}

func (ctrler *AcctCtrler) Commit() ([]byte, int64, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.acctLedger.Commit()
}

func (ctrler *AcctCtrler) Close() xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if ctrler.acctLedger != nil {
		if xerr := ctrler.acctLedger.Close(); xerr != nil {
			ctrler.logger.Error("AcctCtrler", "acctLedger.Close() returns error", xerr.Error())
		}
		ctrler.acctLedger = nil
	}
	return nil
}

func (ctrler *AcctCtrler) findAccount(addr types.Address, exec bool) *atypes.Account {
	k := ledger.ToLedgerKey(addr)

	fn := ctrler.acctLedger.Get
	if exec {
		fn = ctrler.acctLedger.GetFinality
	}

	if acct, xerr := fn(k); xerr != nil {
		// todo: xerr is not returned
		return nil
	} else {
		return acct
	}
}

func (ctrler *AcctCtrler) newAccount(addr types.Address) *atypes.Account {
	return atypes.NewAccountWithName(addr, "")
}

func (ctrler *AcctCtrler) FindOrNewAccount(addr types.Address, exec bool) *atypes.Account {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// `AcctCtrler` MUST be locked until new account is set to acctLedger (issue #32)

	if acct := ctrler.findAccount(addr, exec); acct != nil {
		return acct
	}

	newAcct := ctrler.newAccount(addr)
	if newAcct != nil {
		ctrler.setAccountCommittable(newAcct, exec)
	}
	return newAcct
}

func (ctrler *AcctCtrler) FindAccount(addr types.Address, exec bool) *atypes.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findAccount(addr, exec)
}

func (ctrler *AcctCtrler) ReadAccount(addr types.Address) *atypes.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.readAccount(addr)
}

func (ctrler *AcctCtrler) readAccount(addr types.Address) *atypes.Account {
	if acct, xerr := ctrler.acctLedger.Read(addr.Array32()); xerr != nil {
		// db error or not found
		return atypes.NewAccount(addr)
	} else {
		return acct
	}
}

func (ctrler *AcctCtrler) Transfer(from, to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	acct0 := ctrler.findAccount(from, exec)
	if acct0 == nil {
		return xerrors.ErrNotFoundAccount
	}
	acct1 := ctrler.findAccount(to, exec)
	if acct1 == nil {
		acct1 = ctrler.newAccount(to)
	}
	xerr := ctrler.transfer(acct0, acct1, amt)
	if xerr != nil {
		return xerr
	}

	if xerr := ctrler.setAccountCommittable(acct0, exec); xerr != nil {
		return xerr
	}
	if xerr := ctrler.setAccountCommittable(acct1, exec); xerr != nil {
		return xerr
	}
	return nil
}

func (ctrler *AcctCtrler) transfer(from, to *atypes.Account, amt *uint256.Int) xerrors.XError {
	if err := from.SubBalance(amt); err != nil {
		return err
	}
	if err := to.AddBalance(amt); err != nil {
		_ = from.AddBalance(amt) // refund
		return err
	}
	return nil
}

func (ctrler *AcctCtrler) Reward(to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	if acct := ctrler.findAccount(to, exec); acct == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := acct.AddBalance(amt); xerr != nil {
		return xerr
	} else if xerr := ctrler.setAccountCommittable(acct, exec); xerr != nil {
		return xerr
	}
	return nil
}

func (ctrler *AcctCtrler) SetAccountCommittable(acct *atypes.Account, exec bool) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.setAccountCommittable(acct, exec)
}

func (ctrler *AcctCtrler) setAccountCommittable(acct *atypes.Account, exec bool) xerrors.XError {
	fn := ctrler.acctLedger.Set
	if exec {
		fn = ctrler.acctLedger.SetFinality
	}
	return fn(acct)
}

var _ atypes.ILedgerHandler = (*AcctCtrler)(nil)
var _ atypes.ITrxHandler = (*AcctCtrler)(nil)
var _ atypes.IBlockHandler = (*AcctCtrler)(nil)
var _ atypes.IAccountHandler = (*AcctCtrler)(nil)
