package account

import (
	"bytes"
	"fmt"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	atypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/genesis"
	"github.com/rigochain/rigo-go/ledger"
	"github.com/rigochain/rigo-go/types"
	abytes "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/big"
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
			logger:     logger,
		}, nil
	}
}

func (ctrler *AcctCtrler) InitLedger(req interface{}) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	genAppState, ok := req.(*genesis.GenesisAppState)
	if !ok {
		return xerrors.New("wrong parameter: AcctCtrler::InitLedger requires *genesis.GenesisAppState")
	}

	for _, holder := range genAppState.AssetHolders {
		addr := append(holder.Address, nil...)
		if bal, ok := new(big.Int).SetString(holder.Balance, 10); !ok {
			return xerrors.New("wrong balance in genesis")
		} else {
			acct := &atypes.Account{
				Address: addr,
				Balance: bal,
			}
			if xerr := ctrler.setAccountCommittable(acct, true); xerr != nil {
				return xerr
			}
		}
	}
	return nil
}

func (ctrler *AcctCtrler) ValidateTrx(ctx *atypes.TrxContext) xerrors.XError {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	// check signature
	var fromAddr types.Address
	var pubBytes abytes.HexBytes

	if ctx.Exec {
		tx := ctx.Tx
		sig := tx.Sig
		tx.Sig = nil
		_txbz, xerr := tx.Encode()
		if xerr != nil {
			return xerr
		}
		if fromAddr, pubBytes, xerr = crypto.Sig2Addr(_txbz, sig); xerr != nil {
			return xerr
		}
		if bytes.Compare(fromAddr, tx.From) != 0 {
			return xerrors.ErrInvalidTrxSig.Wrap(fmt.Errorf("wrong address or sig - expected: %v, actual: %v", tx.From, fromAddr))
		}
	} else {
		fromAddr = ctx.Tx.From
	}

	if acct := ctrler.findAccount(fromAddr, ctx.Exec); acct == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := acct.CheckBalance(ctx.NeedAmt); xerr != nil {
		return xerr
	} else if xerr := acct.CheckNonce(ctx.Tx.Nonce); xerr != nil {
		return xerr
	} else {
		ctx.Sender = acct
		ctx.SenderPubKey = pubBytes
	}
	return nil
}

func (ctrler *AcctCtrler) ExecuteTrx(ctx *atypes.TrxContext) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	// amount + fee
	if xerr := ctx.Sender.SubBalance(ctx.NeedAmt); xerr != nil {
		return xerr
	}

	var receiver *atypes.Account
	if ctx.Tx.Type == atypes.TRX_TRANSFER {
		receiver = ctrler.findOrNewAccount(ctx.Tx.To, ctx.Exec)
		if xerr := receiver.AddBalance(ctx.Tx.Amount); xerr != nil {
			return xerr
		}
	}

	// increase sender's nonce
	ctx.Sender.AddNonce()
	// set used gas
	ctx.GasUsed = ctx.Tx.Gas

	_ = ctrler.setAccountCommittable(ctx.Sender, ctx.Exec)
	if receiver != nil {
		_ = ctrler.setAccountCommittable(receiver, ctx.Exec)
	}

	return nil
}

func (ctrler *AcctCtrler) ValidateBlock(ctx *atypes.BlockContext) xerrors.XError {
	// do nothing
	return nil
}

func (ctrler *AcctCtrler) ExecuteBlock(ctx *atypes.BlockContext) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.logger.Debug("AcctCtrler-ExecuteBlock", "height", ctx.BlockInfo.Header.Height)

	if ctx.BlockInfo.Header.ProposerAddress != nil && ctx.Fee.Sign() > 0 {
		// give fee to block proposer
		if acct, xerr := ctrler.acctLedger.GetFinality(ledger.ToLedgerKey(ctx.BlockInfo.Header.ProposerAddress)); xerr != nil {
			return xerr
		} else if xerr := acct.AddBalance(ctx.Fee); xerr != nil {
			return xerr
		} else {
			return ctrler.setAccountCommittable(acct, true)
		}
	}
	return nil
}

func (ctrler *AcctCtrler) Commit() ([]byte, int64, xerrors.XError) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	ctrler.logger.Debug("AcctCtrler-Commit")

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

func (ctrler *AcctCtrler) findOrNewAccount(addr types.Address, exec bool) *atypes.Account {
	if acct := ctrler.findAccount(addr, exec); acct != nil {
		return acct
	}
	return atypes.NewAccountWithName(addr, "")
}

func (ctrler *AcctCtrler) FindAccount_ExecLedger(addr types.Address) *atypes.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findAccount(addr, true)
}

func (ctrler *AcctCtrler) FindAccount_SimuLedger(addr types.Address) *atypes.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findAccount(addr, false)
}

func (ctrler *AcctCtrler) FindOrNewAccount(addr types.Address, exec bool) *atypes.Account {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	return ctrler.findOrNewAccount(addr, exec)
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

func (ctrler *AcctCtrler) Transfer(from, to types.Address, amt *big.Int, exec bool) xerrors.XError {
	ctrler.mtx.RLock()
	defer ctrler.mtx.RUnlock()

	if acct0 := ctrler.findAccount(from, exec); acct0 == nil {
		return xerrors.ErrNotFoundAccount
	} else if acct1 := ctrler.findOrNewAccount(to, exec); acct1 == nil {
		// not reachable
		return xerrors.ErrNotFoundAccount
	} else if xerr := ctrler.transfer(acct0, acct1, amt); xerr != nil {
		return xerr
	} else if xerr := ctrler.setAccountCommittable(acct0, exec); xerr != nil {
		return xerr
	} else if xerr := ctrler.setAccountCommittable(acct1, exec); xerr != nil {
		return xerr
	}
	return nil
}

func (ctrler *AcctCtrler) transfer(from, to *atypes.Account, amt *big.Int) xerrors.XError {
	if err := from.SubBalance(amt); err != nil {
		return err
	}
	if err := to.AddBalance(amt); err != nil {
		_ = from.AddBalance(amt) // refund
		return err
	}
	return nil
}

func (ctrler *AcctCtrler) Reward(to types.Address, amt *big.Int, exec bool) xerrors.XError {
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

func (ctrler *AcctCtrler) SetAccountCommittable(acct *atypes.Account) xerrors.XError {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()

	return ctrler.setAccountCommittable(acct, true)
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
var _ atypes.IAccountHelper = (*AcctCtrler)(nil)
