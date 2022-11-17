package stake_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/account"
	"github.com/kysee/arcanus/types/trxs"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/big"
	"math/rand"
)

type accountHandler struct{}

func (a *accountHandler) FindOrNewAccount(addr account.Address, b bool) *account.Account {
	panic("Don't use this method")
}

func (a *accountHandler) FindAccount(addr account.Address, b bool) *account.Account {
	if w := FindWallet(addr); w != nil {
		return w.GetAccount()
	}
	return nil
}

var _ account.IAccountFinder = (*accountHandler)(nil)

func FindWallet(addr account.Address) *client.Wallet {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.Address()) == 0 {
			return w
		}
	}
	return nil
}

type govRuleHandler struct{}

var (
	amtPerPower    = big.NewInt(1000)
	rewardPerPower = big.NewInt(10)
)

func (g govRuleHandler) GetMaxValidatorCount() int64 {
	return 21
}

func (g govRuleHandler) GetLazyRewardBlocks() int64 {
	return 10
}

func (g govRuleHandler) GetLazyApplyingBlocks() int64 {
	return 10
}

func (g govRuleHandler) MaxStakeAmount() *big.Int {
	return new(big.Int).Mul(big.NewInt(tmtypes.MaxTotalVotingPower), amtPerPower)
}

func (g govRuleHandler) MaxTotalPower() int64 {
	return tmtypes.MaxTotalVotingPower
}

func (g govRuleHandler) AmountToPower(amt *big.Int) int64 {
	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, amtPerPower)
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (g govRuleHandler) PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), amtPerPower)
}

func (g govRuleHandler) PowerToReward(power int64) *big.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), rewardPerPower)
}

var _ types.IGovRuleHandler = (*govRuleHandler)(nil)

func makeTestWallets(n int) []*client.Wallet {
	wallets := make([]*client.Wallet, n)
	for i := 0; i < n; i++ {
		w := client.NewWallet([]byte("1"))
		w.GetAccount().AddBalance(types.ToSAU(100000000))
		w.Unlock([]byte("1"))

		wallets[i] = w
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

func makeStakingTrxContext(from, to *client.Wallet, power int64) (*trxs.TrxContext, error) {
	amt := govRuleHandlerHelper.PowerToAmount(power)

	tx := client.NewTrxStaking(from.Address(), to.Address(), dummyGas, amt, dummyNonce)
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	return &trxs.TrxContext{
		Exec:           true,
		Tx:             tx,
		TxHash:         crypto.DefaultHash(bz),
		Height:         lastHeight,
		SenderPubKey:   from.GetPubKey(),
		Sender:         from.GetAccount(),
		Receiver:       to.GetAccount(),
		NeedAmt:        nil,
		GasUsed:        nil,
		GovRuleHandler: govRuleHandlerHelper,
		Error:          nil,
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

func makeUnstakingTrxContext(from, to *client.Wallet, txhash types.HexBytes) (*trxs.TrxContext, error) {

	tx := client.NewTrxUnstaking(from.Address(), to.Address(), dummyGas, dummyNonce, txhash)
	tzbz, _, err := from.SignTrx(tx)
	if err != nil {
		return nil, err
	}

	return &trxs.TrxContext{
		Exec:           true,
		Tx:             tx,
		TxHash:         crypto.DefaultHash(tzbz),
		Height:         lastHeight,
		SenderPubKey:   from.GetPubKey(),
		Sender:         from.GetAccount(),
		Receiver:       to.GetAccount(),
		GovRuleHandler: govRuleHandlerHelper,
	}, nil
}
