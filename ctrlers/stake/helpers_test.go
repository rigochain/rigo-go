package stake_test

import (
	"bytes"
	"errors"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"math/big"
	"math/rand"
)

type TestWallet struct {
	types.IAccount
	W *client.Wallet
}

type StakingResult struct {
	Owner  types.Address
	To     types.Address
	Amt    *big.Int
	Power  int64
	TxHash types.HexBytes
	Height int64
}

type AccountHelper struct{}

func (a *AccountHelper) FindOrNewAccount(addr types.Address, b bool) types.IAccount {
	panic("Don't use this method")
}

func (a *AccountHelper) FindAccount(addr types.Address, b bool) types.IAccount {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.W.Address()) == 0 {
			return w
		}
	}
	return nil
}

var _ types.IAccountFinder = (*AccountHelper)(nil)

func makeTestWallets(n int) []*TestWallet {
	wallets := make([]*TestWallet, n)
	for i := 0; i < n; i++ {
		w := client.NewWallet([]byte("1"))
		w.Unlock([]byte("1"))
		a := account.NewAccount(w.Address())
		a.AddBalance(types.ToSAU(100000000))

		wallets[i] = &TestWallet{a, w}
	}
	return wallets
}

func randMakeStakingToSelfTrxContext() (*trxs.TrxContext, error) {
	from := Wallets[rand.Intn(len(Wallets))]
	to := from

	power := libs.RandInt63n(1000)

	if txCtx, err := makeStakingTrxContext(from, to, power); err != nil {
		return nil, err
	} else {
		DelegateeWallets = append(DelegateeWallets, from)
		return txCtx, nil
	}

}

func randMakeStakingTrxContext() (*trxs.TrxContext, error) {
	from, to := Wallets[rand.Intn(len(Wallets))], DelegateeWallets[rand.Intn(len(DelegateeWallets))]
	power := libs.RandInt63n(1000)
	return makeStakingTrxContext(from, to, power)
}

func makeStakingTrxContext(from, to *TestWallet, power int64) (*trxs.TrxContext, error) {
	amt := testGovRules.PowerToAmount(power)

	tx := client.NewTrxStaking(from.W.Address(), to.W.Address(), dummyGas, amt, dummyNonce)
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	return &trxs.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(bz),
		Height:       lastHeight,
		SenderPubKey: from.W.GetPubKey(),
		Sender:       from,
		Receiver:     to,
		NeedAmt:      nil,
		GasUsed:      nil,
		GovRules:     testGovRules,
		Error:        nil,
	}, nil
}

func findStakingTxCtx(txhash types.HexBytes) *trxs.TrxContext {
	for _, tctx := range stakingTrxCtxs {
		if bytes.Compare(tctx.TxHash, txhash) == 0 {
			return tctx
		}
	}
	return nil
}

func randMakeUnstakingTrxContext() (*trxs.TrxContext, error) {
	rn := rand.Intn(len(stakingTrxCtxs))
	stakingTxCtx := stakingTrxCtxs[rn]

	from := acctCtrlerHelper.FindAccount(stakingTxCtx.Tx.From, true)
	if from == nil {
		return nil, errors.New("not found test account for " + stakingTxCtx.Tx.From.String())
	}
	to := acctCtrlerHelper.FindAccount(stakingTxCtx.Tx.To, true)
	if to == nil {
		return nil, errors.New("not found test account for " + stakingTxCtx.Tx.To.String())
	}

	return makeUnstakingTrxContext(from.(*TestWallet), to.(*TestWallet), stakingTxCtx.TxHash)
}

func makeUnstakingTrxContext(from, to *TestWallet, txhash types.HexBytes) (*trxs.TrxContext, error) {

	tx := client.NewTrxUnstaking(from.W.Address(), to.W.Address(), dummyGas, dummyNonce, txhash)
	tzbz, _, err := from.W.SignTrx(tx)
	if err != nil {
		return nil, err
	}

	return &trxs.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(tzbz),
		Height:       lastHeight,
		SenderPubKey: from.W.GetPubKey(),
		Sender:       from,
		Receiver:     to,
		GovRules:     testGovRules,
	}, nil
}
