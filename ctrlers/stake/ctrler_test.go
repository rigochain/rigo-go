package stake_test

import (
	"bytes"
	"github.com/holiman/uint256"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	types2 "github.com/tendermint/tendermint/proto/tendermint/types"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

var (
	config      = cfg.DefaultConfig()
	DBDIR       = filepath.Join(os.TempDir(), "stake-ctrler-unstaking-test")
	acctHelper  = &acctHandlerMock{}
	govHelper   = &govHandlerMock{}
	stakeCtrler *stake.StakeCtrler

	Wallets              []*web3.Wallet
	DelegateeWallets     []*web3.Wallet
	stakingToSelfTrxCtxs []*types.TrxContext
	stakingTrxCtxs       []*types.TrxContext
	unstakingTrxCtxs     []*types.TrxContext

	dummyGas   = uint256.NewInt(0)
	dummyNonce = uint64(0)

	lastHeight = int64(1)
)

func TestMain(m *testing.M) {
	os.RemoveAll(DBDIR)

	config.DBPath = DBDIR
	stakeCtrler, _ = stake.NewStakeCtrler(config, govHelper, tmlog.NewNopLogger())

	Wallets = makeTestWallets(rand.Intn(100) + int(govHelper.MaxValidatorCnt()))

	for i := 0; i < 5; i++ {
		if txctx, err := randMakeStakingToSelfTrxContext(); err != nil {
			panic(err)
		} else {
			stakingToSelfTrxCtxs = append(stakingToSelfTrxCtxs, txctx)
		}
		if rand.Int()%3 == 0 {
			lastHeight++
		}
	}

	for i := 0; i < 1000; i++ {
		if txctx, err := randMakeStakingTrxContext(); err != nil {
			panic(err)
		} else {
			stakingTrxCtxs = append(stakingTrxCtxs, txctx)
		}
		if rand.Int()%3 == 0 {
			lastHeight++
		}
	}

	lastHeight += 10

	for i := 0; i < 100; i++ {
		if txctx, err := randMakeUnstakingTrxContext(); err != nil {
			panic(err)
		} else {
			already := false
			for _, _ctx := range unstakingTrxCtxs {
				if bytes.Compare(_ctx.Tx.Payload.(*types.TrxPayloadUnstaking).TxHash, txctx.Tx.Payload.(*types.TrxPayloadUnstaking).TxHash) == 0 {
					already = true
				}
			}
			if !already {
				unstakingTrxCtxs = append(unstakingTrxCtxs, txctx)
			}

		}
		if rand.Int()%3 == 0 {
			lastHeight++
		}
	}

	exitCode := m.Run()

	os.RemoveAll(DBDIR)

	os.Exit(exitCode)
}

func TestStakingToSelfByTx(t *testing.T) {
	sumAmt := uint256.NewInt(0)
	sumPower := int64(0)

	for _, txctx := range stakingToSelfTrxCtxs {
		err := stakeCtrler.ExecuteTrx(txctx)
		require.NoError(t, err)

		_ = sumAmt.Add(sumAmt, txctx.Tx.Amount)
		sumPower += txctx.GovHandler.AmountToPower(txctx.Tx.Amount)
	}

	_, _, err := stakeCtrler.Commit()
	require.NoError(t, err)

	require.Equal(t, sumAmt.String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.GetTotalPower())
}

func TestStakingByTx(t *testing.T) {
	sumAmt := stakeCtrler.GetTotalAmount()
	sumPower := stakeCtrler.GetTotalPower()

	for _, txctx := range stakingTrxCtxs {
		err := stakeCtrler.ExecuteTrx(txctx)
		require.NoError(t, err)

		_ = sumAmt.Add(sumAmt, txctx.Tx.Amount)
		sumPower += txctx.GovHandler.AmountToPower(txctx.Tx.Amount)
	}

	_, _, err := stakeCtrler.Commit()
	require.NoError(t, err)

	require.Equal(t, sumAmt.String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.GetTotalPower())
}

func TestUnstakingByTx(t *testing.T) {
	sumAmt0 := stakeCtrler.GetTotalAmount()
	sumPower0 := stakeCtrler.GetTotalPower()
	sumUnstakingAmt := uint256.NewInt(0)
	sumUnstakingPower := int64(0)

	for _, txctx := range unstakingTrxCtxs {
		stakingTxHash := txctx.Tx.Payload.(*types.TrxPayloadUnstaking).TxHash

		err := stakeCtrler.ExecuteTrx(txctx)
		require.NoError(t, err)

		stakingTxCtx := findStakingTxCtx(stakingTxHash)

		sumUnstakingAmt.Add(sumUnstakingAmt, stakingTxCtx.Tx.Amount)
		sumUnstakingPower += txctx.GovHandler.AmountToPower(stakingTxCtx.Tx.Amount)
	}

	_, _, err := stakeCtrler.Commit()
	require.NoError(t, err)

	require.Equal(t, new(uint256.Int).Sub(sumAmt0, sumUnstakingAmt).String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower0-sumUnstakingPower, stakeCtrler.GetTotalPower())

	// test freezing reward
	frozenStakes := stakeCtrler.GetFrozenStakes()
	require.Equal(t, len(unstakingTrxCtxs), len(frozenStakes))

	sumFrozenAmount := uint256.NewInt(0)
	sumFrozenPower := int64(0)
	for _, s := range frozenStakes {
		sumFrozenAmount.Add(sumFrozenAmount, s.Amount)
		sumFrozenPower += s.Power
	}
	require.Equal(t, sumFrozenAmount.String(), sumUnstakingAmt.String())
	require.Equal(t, sumFrozenPower, sumUnstakingPower)
}

func TestUnfreezing(t *testing.T) {
	lastHeight += govHelper.LazyRewardBlocks()

	// execute block at lastHeight
	req := abcitypes.RequestBeginBlock{
		Header: types2.Header{
			Height: lastHeight,
		},
	}
	bctx := types.NewBlockContext(req, govHelper, acctHelper, nil)
	bctx.AddGas(uint256.NewInt(10))
	err := stakeCtrler.ExecuteBlock(bctx)
	require.NoError(t, err)

	_, _, err = stakeCtrler.Commit()
	require.NoError(t, err)

	frozenStakes := stakeCtrler.GetFrozenStakes()
	require.Equal(t, 0, len(frozenStakes))

	// todo: check received reward and fee
}
