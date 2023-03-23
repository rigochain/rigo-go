package stake_test

import (
	"bytes"
	"errors"
	"fmt"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/big"
	"math/rand"
)

type accountHandler struct{}

func (a *accountHandler) FindOrNewAccount(addr types.Address, exec bool) *types2.Account {
	panic("Don't use this method")
}

func (a *accountHandler) FindAccount(addr types.Address, exec bool) *types2.Account {
	if w := FindWallet(addr); w != nil {
		return w.GetAccount()
	}
	return nil
}
func (a *accountHandler) Transfer(from, to types.Address, amt *big.Int, exec bool) xerrors.XError {
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
func (a *accountHandler) Reward(to types.Address, amt *big.Int, exec bool) xerrors.XError {
	if receiver := a.FindAccount(to, exec); receiver == nil {
		return xerrors.ErrNotFoundAccount
	} else if xerr := receiver.AddBalance(amt); xerr != nil {
		return xerr
	}
	return nil
}

var _ types2.IAccountHelper = (*accountHandler)(nil)

func FindWallet(addr types.Address) *web3.Wallet {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.Address()) == 0 {
			return w
		}
	}
	return nil
}

type govHelperMock struct{}

func (g *govHelperMock) Version() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHelperMock) AmountPerPower() *big.Int {
	return big.NewInt(1000)
}

func (g *govHelperMock) RewardPerPower() *big.Int {
	return big.NewInt(10)
}

func (g *govHelperMock) MinTrxFee() *big.Int {
	//TODO implement me
	panic("implement me")
}

func (g *govHelperMock) MinVotingPeriodBlocks() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHelperMock) MaxVotingPeriodBlocks() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHelperMock) MaxValidatorCnt() int64 {
	return 21
}

func (g *govHelperMock) LazyRewardBlocks() int64 {
	return 10
}

func (g *govHelperMock) LazyApplyingBlocks() int64 {
	return 10
}

func (g *govHelperMock) MaxStakeAmount() *big.Int {
	return new(big.Int).Mul(big.NewInt(tmtypes.MaxTotalVotingPower), g.AmountPerPower())
}

func (g *govHelperMock) MaxTotalPower() int64 {
	return tmtypes.MaxTotalVotingPower
}

func (g *govHelperMock) AmountToPower(amt *big.Int) int64 {
	// 1 VotingPower == 1 XCO
	_vp := new(big.Int).Quo(amt, g.AmountPerPower())
	vp := _vp.Int64()
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (g *govHelperMock) PowerToAmount(power int64) *big.Int {
	// 1 VotingPower == 1 XCO
	return new(big.Int).Mul(big.NewInt(power), g.AmountPerPower())
}

func (g *govHelperMock) PowerToReward(power int64) *big.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	return new(big.Int).Mul(big.NewInt(power), g.RewardPerPower())
}

var _ types2.IGovHelper = (*govHelperMock)(nil)

func makeTestWallets(n int) []*web3.Wallet {
	wallets := make([]*web3.Wallet, n)
	for i := 0; i < n; i++ {
		w := web3.NewWallet([]byte("1"))
		w.GetAccount().AddBalance(types.ToSAU(100000000))
		w.Unlock([]byte("1"))

		wallets[i] = w
	}
	return wallets
}

func randMakeStakingToSelfTrxContext() (*types2.TrxContext, error) {
	from := Wallets[rand.Intn(len(Wallets))]
	to := from

	power := bytes2.RandInt63n(1000)

	if txCtx, err := makeStakingTrxContext(from, to, power); err != nil {
		return nil, err
	} else {
		DelegateeWallets = append(DelegateeWallets, from)
		return txCtx, nil
	}

}

func randMakeStakingTrxContext() (*types2.TrxContext, error) {
	from, to := Wallets[rand.Intn(len(Wallets))], DelegateeWallets[rand.Intn(len(DelegateeWallets))]
	power := bytes2.RandInt63n(1000)
	return makeStakingTrxContext(from, to, power)
}

func makeStakingTrxContext(from, to *web3.Wallet, power int64) (*types2.TrxContext, error) {
	amt := govHelper.PowerToAmount(power)

	tx := web3.NewTrxStaking(from.Address(), to.Address(), dummyNonce, dummyGas, amt)
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	return &types2.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(bz),
		Height:       lastHeight,
		SenderPubKey: from.GetPubKey(),
		Sender:       from.GetAccount(),
		Receiver:     to.GetAccount(),
		NeedAmt:      nil,
		GasUsed:      nil,
		GovHelper:    govHelper,
	}, nil
}

func findStakingTxCtx(txhash bytes2.HexBytes) *types2.TrxContext {
	for _, tctx := range stakingTrxCtxs {
		if bytes.Compare(tctx.TxHash, txhash) == 0 {
			return tctx
		}
	}
	return nil
}

func randMakeUnstakingTrxContext() (*types2.TrxContext, error) {
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

func makeUnstakingTrxContext(from, to *web3.Wallet, txhash bytes2.HexBytes) (*types2.TrxContext, error) {

	tx := web3.NewTrxUnstaking(from.Address(), to.Address(), dummyNonce, dummyGas, txhash)
	tzbz, _, err := from.SignTrx(tx)
	if err != nil {
		return nil, err
	}

	return &types2.TrxContext{
		Exec:         true,
		Tx:           tx,
		TxHash:       crypto.DefaultHash(tzbz),
		Height:       lastHeight,
		SenderPubKey: from.GetPubKey(),
		Sender:       from.GetAccount(),
		Receiver:     to.GetAccount(),
		GovHelper:    govHelper,
	}, nil
}
