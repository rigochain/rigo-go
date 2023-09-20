package stake_test

import (
	"bytes"
	"errors"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	"math/rand"
)

type acctHandlerMock struct{}

func (a *acctHandlerMock) FindOrNewAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	panic("Don't use this method")
}

func (a *acctHandlerMock) FindAccount(addr types.Address, exec bool) *ctrlertypes.Account {
	if w := FindWallet(addr); w != nil {
		return w.GetAccount()
	}
	return nil
}
func (a *acctHandlerMock) Transfer(from, to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	if sender := a.FindAccount(from, exec); sender == nil {
		return xerrors.ErrNotFoundAccount
	} else if receiver := a.FindAccount(to, exec); receiver == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := sender.SubBalance(amt); xerr != nil {
		return xerr
	} else if xerr := receiver.AddBalance(amt); xerr != nil {
		return xerr
	}
	return nil
}
func (a *acctHandlerMock) Reward(to types.Address, amt *uint256.Int, exec bool) xerrors.XError {
	if receiver := a.FindAccount(to, exec); receiver == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := receiver.AddBalance(amt); xerr != nil {
		return xerr
	}
	return nil
}

func (a *acctHandlerMock) ImmutableAcctCtrlerAt(i int64) (ctrlertypes.IAccountHandler, xerrors.XError) {
	return &acctHandlerMock{}, nil
}

func (a *acctHandlerMock) SetAccountCommittable(account *ctrlertypes.Account, b bool) xerrors.XError {
	return nil
}

var _ ctrlertypes.IAccountHandler = (*acctHandlerMock)(nil)

func FindWallet(addr types.Address) *web3.Wallet {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.Address()) == 0 {
			return w
		}
	}
	return nil
}

func makeTestWallets(n int) []*web3.Wallet {
	wallets := make([]*web3.Wallet, n)
	for i := 0; i < n; i++ {
		w := web3.NewWallet([]byte("1"))
		w.GetAccount().AddBalance(types.ToFons(100000000))
		w.Unlock([]byte("1"))

		wallets[i] = w
	}
	return wallets
}

func randMakeStakingToSelfTrxContext() (*ctrlertypes.TrxContext, error) {
	from := Wallets[rand.Intn(len(Wallets))]
	to := from

	power := ctrlertypes.AmountToPower(govParams.MinValidatorStake()) + rand.Int63n(10000)

	if txCtx, err := makeStakingTrxContext(from, to, power); err != nil {
		return nil, err
	} else {
		DelegateeWallets = append(DelegateeWallets, to)
		return txCtx, nil
	}

}

func randMakeStakingTrxContext() (*ctrlertypes.TrxContext, error) {
	for {
		from, to := Wallets[rand.Intn(len(Wallets))], DelegateeWallets[rand.Intn(len(DelegateeWallets))]
		if bytes.Compare(from.Address(), to.Address()) == 0 {
			continue
		}
		power := rand.Int63n(1000) + 10
		return makeStakingTrxContext(from, to, power)
	}
}

func makeStakingTrxContext(from, to *web3.Wallet, power int64) (*ctrlertypes.TrxContext, error) {
	amt := ctrlertypes.PowerToAmount(power)

	tx := web3.NewTrxStaking(from.Address(), to.Address(), dummyNonce, dummyGas, dummyGasPrice, amt)
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	return &ctrlertypes.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(bz),
		Height:       lastHeight,
		SenderPubKey: from.GetPubKey(),
		Sender:       from.GetAccount(),
		Receiver:     to.GetAccount(),
		GasUsed:      0,
		GovHandler:   govParams,
		AcctHandler:  acctHelper,
	}, nil
}

func findStakingTxCtx(txhash bytes2.HexBytes) *ctrlertypes.TrxContext {
	for _, tctx := range stakingTrxCtxs {
		if bytes.Compare(tctx.TxHash, txhash) == 0 {
			return tctx
		}
	}
	return nil
}

func randMakeUnstakingTrxContext() (*ctrlertypes.TrxContext, error) {
	rn := rand.Intn(len(stakingTrxCtxs))
	stakingTxCtx := stakingTrxCtxs[rn]

	from := FindWallet(stakingTxCtx.Tx.From)
	if from == nil {
		return nil, errors.New("not found test account for " + stakingTxCtx.Tx.From.String())
	}
	to := FindWallet(stakingTxCtx.Tx.To)
	if to == nil {
		return nil, errors.New("not found test account for " + stakingTxCtx.Tx.To.String())
	}

	return makeUnstakingTrxContext(from, to, stakingTxCtx.TxHash)
}

func makeUnstakingTrxContext(from, to *web3.Wallet, txhash bytes2.HexBytes) (*ctrlertypes.TrxContext, error) {

	tx := web3.NewTrxUnstaking(from.Address(), to.Address(), dummyNonce, dummyGas, dummyGasPrice, txhash)
	tzbz, _, err := from.SignTrxRLP(tx, "stake_test_chain")
	if err != nil {
		return nil, err
	}

	return &ctrlertypes.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(tzbz),
		Height:       lastHeight,
		SenderPubKey: from.GetPubKey(),
		Sender:       from.GetAccount(),
		Receiver:     to.GetAccount(),
		GovHandler:   govParams,
	}, nil
}
