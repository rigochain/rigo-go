package stake_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/holiman/uint256"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	bytes2 "github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/rand"
)

type acctHandlerMock struct{}

func (a *acctHandlerMock) FindOrNewAccount(addr types.Address, exec bool) *types2.Account {
	panic("Don't use this method")
}

func (a *acctHandlerMock) FindAccount(addr types.Address, exec bool) *types2.Account {
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

func (a *acctHandlerMock) ImmutableAcctCtrlerAt(i int64) (types2.IAccountHandler, xerrors.XError) {
	return &acctHandlerMock{}, nil
}

func (a *acctHandlerMock) SetAccountCommittable(account *types2.Account, b bool) xerrors.XError {
	return nil
}

var _ types2.IAccountHandler = (*acctHandlerMock)(nil)

func FindWallet(addr types.Address) *web3.Wallet {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.Address()) == 0 {
			return w
		}
	}
	return nil
}

type govHandlerMock struct{}

func (g *govHandlerMock) Version() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHandlerMock) AmountPerPower() *uint256.Int {
	return uint256.NewInt(1_000_000_000_000_000_000)
}

func (g *govHandlerMock) RewardPerPower() *uint256.Int {
	return uint256.NewInt(10)
}

func (g *govHandlerMock) MinTrxFee() *uint256.Int {
	//TODO implement me
	panic("implement me")
}

func (g *govHandlerMock) MinVotingPeriodBlocks() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHandlerMock) MaxVotingPeriodBlocks() int64 {
	//TODO implement me
	panic("implement me")
}

func (g *govHandlerMock) MaxValidatorCnt() int64 {
	return 21
}

func (g *govHandlerMock) LazyRewardBlocks() int64 {
	return 10
}

func (g *govHandlerMock) LazyApplyingBlocks() int64 {
	return 10
}

func (g *govHandlerMock) MaxStakeAmount() *uint256.Int {
	return new(uint256.Int).Mul(uint256.NewInt(uint64(tmtypes.MaxTotalVotingPower)), g.AmountPerPower())
}

func (g *govHandlerMock) MaxTotalPower() int64 {
	return tmtypes.MaxTotalVotingPower
}

func (g *govHandlerMock) MinSelfStakeRatio() int64 {
	return 50
}

func (g *govHandlerMock) MaxUpdatableStakeRatio() int64 {
	return 30
}

func (g *govHandlerMock) SlashRatio() int64 {
	return 27
}

func (g *govHandlerMock) AmountToPower(amt *uint256.Int) int64 {
	// 1 VotingPower == 1_000_000_000_000_000_000 (10^18)
	_vp := new(uint256.Int).Div(amt, g.AmountPerPower())
	vp := int64(_vp.Uint64())
	if vp < 0 {
		panic(fmt.Sprintf("voting power is negative: %v", vp))
	}
	return vp
}

func (g *govHandlerMock) PowerToAmount(power int64) *uint256.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	// 1 VotingPower == 1 XCO
	_power := uint64(power)
	return new(uint256.Int).Mul(uint256.NewInt(_power), g.AmountPerPower())
}

func (g *govHandlerMock) PowerToReward(power int64) *uint256.Int {
	if power < 0 {
		panic(fmt.Sprintf("power is negative: %v", power))
	}
	_power := uint64(power)
	return new(uint256.Int).Mul(uint256.NewInt(_power), g.RewardPerPower())
}

var _ types2.IGovHandler = (*govHandlerMock)(nil)

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

	power := rand.Int63n(1000)

	if txCtx, err := makeStakingTrxContext(from, to, power); err != nil {
		return nil, err
	} else {
		DelegateeWallets = append(DelegateeWallets, from)
		return txCtx, nil
	}

}

func randMakeStakingTrxContext() (*types2.TrxContext, error) {
	from, to := Wallets[rand.Intn(len(Wallets))], DelegateeWallets[rand.Intn(len(DelegateeWallets))]
	power := rand.Int63n(1000)
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
		GovHandler:   govHelper,
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
		GovHandler:   govHelper,
	}, nil
}
