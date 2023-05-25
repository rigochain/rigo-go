package stake_test

import (
	"bytes"
	"github.com/holiman/uint256"
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	types3 "github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	types2 "github.com/tendermint/tendermint/proto/tendermint/types"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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

	Wallets = makeTestWallets(100 + int(govHelper.MaxValidatorCnt()))

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

func TestTrxStakingToSelf(t *testing.T) {
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

	require.Equal(t, sumAmt.String(), stakeCtrler.ReadTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.ReadTotalPower())
}

func TestTrxStakingByTx(t *testing.T) {
	sumAmt := stakeCtrler.ReadTotalAmount()
	sumPower := stakeCtrler.ReadTotalPower()

	for _, txctx := range stakingTrxCtxs {
		power0 := stakeCtrler.SelfPowerOf(txctx.Tx.To)
		power1 := stakeCtrler.DelegatedPowerOf(txctx.Tx.To)
		maxAmt := govHelper.PowerToAmount(power0 - power1)

		err := stakeCtrler.ExecuteTrx(txctx)

		if txctx.Tx.Amount.Cmp(maxAmt) > 0 && txctx.Tx.From.Compare(txctx.Tx.To) != 0 {
			// it's error to try delegating to validator over self_stake_ratio
			require.Error(t, err)
			for i, ctx := range unstakingTrxCtxs {
				if bytes.Compare(ctx.Tx.Payload.(*types.TrxPayloadUnstaking).TxHash, txctx.TxHash) == 0 {
					unstakingTrxCtxs = append(unstakingTrxCtxs[:i], unstakingTrxCtxs[i+1:]...)
					break
				}
			}
		} else {
			require.NoError(t, err)
			_ = sumAmt.Add(sumAmt, txctx.Tx.Amount)
			sumPower += txctx.GovHandler.AmountToPower(txctx.Tx.Amount)
		}
	}

	_, _, err := stakeCtrler.Commit()
	require.NoError(t, err)

	require.Equal(t, sumAmt.String(), stakeCtrler.ReadTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.ReadTotalPower())
}

func TestDoReward(t *testing.T) {
	valUps := stakeCtrler.UpdateValidators(int(govHelper.MaxValidatorCnt()))
	require.Greater(t, len(valUps), 0)

	var votes []abcitypes.VoteInfo
	for _, val := range valUps {
		addr, err := crypto.PubBytes2Addr(val.PubKey.GetSecp256K1())
		require.NoError(t, err)

		votes = append(votes, abcitypes.VoteInfo{
			Validator: abcitypes.Validator{
				Address: addr,
				Power:   val.Power,
			},
			SignedLastBlock: true,
		})
	}
	lastHeight++
	err := stakeCtrler.DoReward(lastHeight, votes)
	require.NoError(t, err)
	_, _, err = stakeCtrler.Commit()
	require.NoError(t, err)
}

func TestPunish(t *testing.T) {
	allTotalPower0 := stakeCtrler.ReadTotalPower()
	for _, byzantine := range DelegateeWallets {
		totalPower0 := stakeCtrler.PowerOf(byzantine.Address())
		selfPower0 := stakeCtrler.SelfPowerOf(byzantine.Address())
		require.Greater(t, selfPower0, int64(0))

		_slashed := uint256.NewInt(uint64(selfPower0))
		_ = _slashed.Mul(_slashed, uint256.NewInt(uint64(govHelper.SlashRatio())))
		_ = _slashed.Div(_slashed, uint256.NewInt(uint64(100)))
		expectedSlashed := int64(_slashed.Uint64())
		require.Greater(t, expectedSlashed, int64(0))
		require.Greater(t, selfPower0, expectedSlashed)

		// original stake state
		delegatee := stakeCtrler.Delegatee(byzantine.Address())
		oriStakes := make([]*stake.Stake, delegatee.StakesLen())
		for i, s0 := range delegatee.GetAllStakes() {
			oriStakes[i] = s0.Clone()
		}

		ev, xerr := stakeCtrler.BeginBlock(types.NewBlockContext(
			abcitypes.RequestBeginBlock{
				ByzantineValidators: []abcitypes.Evidence{
					{
						Validator: abcitypes.Validator{
							Address: byzantine.Address(),
							Power:   totalPower0,
						},
					},
				},
			}, govHelper, nil, nil),
		)
		require.NoError(t, xerr)
		require.Equal(t, []byte("slashed"), ev[0].Attributes[3].Key)

		totalSlashedPower0, err := strconv.ParseInt(string(ev[0].Attributes[3].Value), 10, 64)
		require.NoError(t, err)

		delegatee = stakeCtrler.Delegatee(byzantine.Address())

		expectedTotalSlashedPower0 := int64(0)
		expectedTotalReceivedReward := uint256.NewInt(0)
		for _, s0 := range oriStakes {
			p0 := uint256.NewInt(uint64(s0.Power))
			_ = p0.Mul(p0, uint256.NewInt(uint64(govHelper.SlashRatio())))
			_ = p0.Div(p0, uint256.NewInt(uint64(100)))
			slashedPower := int64(p0.Uint64())

			if slashedPower < 1 {
				slashedPower = s0.Power
				expectedTotalSlashedPower0 += slashedPower
				continue
			}
			expectedPower := s0.Power - slashedPower
			require.NotEqual(t, expectedPower, s0.Power)

			expectedAmt := new(uint256.Int).Mul(govHelper.AmountPerPower(), uint256.NewInt(uint64(expectedPower)))
			require.NotEqual(t, expectedAmt.Dec(), s0.Amount.Dec())

			_blocks := uint64(0)
			if s0.ReceivedReward.Sign() > 0 {
				_blocks = new(uint256.Int).Div(s0.ReceivedReward, s0.BlockRewardUnit).Uint64()
			}

			expectedBlockRewardUnit := new(uint256.Int).Mul(govHelper.RewardPerPower(), uint256.NewInt(uint64(expectedPower)))
			require.NotEqual(t, expectedBlockRewardUnit.Dec(), s0.BlockRewardUnit.Dec())
			expectedReceivedReward := new(uint256.Int).Mul(expectedBlockRewardUnit, uint256.NewInt(_blocks))
			require.NotEqual(t, expectedReceivedReward.Dec(), s0.ReceivedReward.Dec())

			_, s1 := delegatee.FindStake(s0.TxHash)

			require.NotNil(t, s1)
			require.Equal(t, s0.TxHash, s1.TxHash)
			require.Equal(t, expectedPower, s1.Power)
			require.Equal(t, expectedAmt.Dec(), s1.Amount.Dec())
			require.Equal(t, expectedBlockRewardUnit.Dec(), s1.BlockRewardUnit.Dec())
			require.Equal(t, expectedReceivedReward.Dec(), s1.ReceivedReward.Dec())

			expectedTotalSlashedPower0 += slashedPower
			_ = expectedTotalReceivedReward.Add(expectedTotalReceivedReward, expectedReceivedReward)
		}
		require.Equal(t, expectedTotalSlashedPower0, totalSlashedPower0)
		require.Equal(t, totalPower0-totalSlashedPower0, delegatee.TotalPower)
		require.Equal(t, totalPower0-totalSlashedPower0, delegatee.SumPowerOf(nil))
		require.Equal(t, expectedTotalReceivedReward.Dec(), delegatee.TotalRewardAmount.Dec())

		_, _, xerr = stakeCtrler.Commit()
		require.NoError(t, xerr)

		require.Equal(t, totalPower0-expectedTotalSlashedPower0, stakeCtrler.PowerOf(delegatee.Addr))
		require.Equal(t, allTotalPower0-expectedTotalSlashedPower0, stakeCtrler.ReadTotalPower())

		allTotalPower0 -= expectedTotalSlashedPower0
	}
}

// test for issue #43
func TestUnstakingByNotOwner(t *testing.T) {
	for _, txctx := range unstakingTrxCtxs {
		ori := txctx.Tx.From
		txctx.Tx.From = types3.RandAddress()
		err := stakeCtrler.ExecuteTrx(txctx)
		require.Error(t, err)

		txctx.Tx.From = ori
	}
}

func TestUnstakingByTx(t *testing.T) {
	sumAmt0 := stakeCtrler.ReadTotalAmount()
	sumPower0 := stakeCtrler.ReadTotalPower()
	sumUnstakingAmt := uint256.NewInt(0)
	sumUnstakingPower := int64(0)

	for _, txctx := range unstakingTrxCtxs {
		stakingTxHash := txctx.Tx.Payload.(*types.TrxPayloadUnstaking).TxHash
		delegatee := stakeCtrler.Delegatee(txctx.Tx.To)
		require.NotNil(t, delegatee)

		_, s0 := delegatee.FindStake(stakingTxHash)
		require.NotNil(t, s0)

		err := stakeCtrler.ExecuteTrx(txctx)
		require.NoError(t, err)

		sumUnstakingAmt.Add(sumUnstakingAmt, s0.Amount)
		sumUnstakingPower += txctx.GovHandler.AmountToPower(s0.Amount)
	}

	_, _, err := stakeCtrler.Commit()
	require.NoError(t, err)

	require.Equal(t, new(uint256.Int).Sub(sumAmt0, sumUnstakingAmt).Dec(), stakeCtrler.ReadTotalAmount().Dec())
	require.Equal(t, sumPower0-sumUnstakingPower, stakeCtrler.ReadTotalPower())

	// test freezing reward
	frozenStakes := stakeCtrler.ReadFrozenStakes()
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
	type expectedReward struct {
		addr            types3.Address
		originalBalance *uint256.Int
		rewardedBalance *uint256.Int
	}

	expectedRewards := make(map[string]*expectedReward)
	frozenStakes := stakeCtrler.ReadFrozenStakes()
	require.Greater(t, len(frozenStakes), 0)
	for _, s0 := range frozenStakes {
		//require.NotEqual(t, "0", s0.ReceivedReward.Dec(), "your test may not reward")

		acct := acctHelper.FindAccount(s0.From, true)
		require.NotNil(t, acct)

		er, ok := expectedRewards[acct.Address.String()]
		if !ok {
			er = &expectedReward{
				addr:            acct.Address,
				originalBalance: acct.Balance.Clone(),
				rewardedBalance: uint256.NewInt(0),
			}
			expectedRewards[acct.Address.String()] = er
		}

		_ = er.rewardedBalance.Add(
			er.rewardedBalance,
			new(uint256.Int).Add(s0.Amount, s0.ReceivedReward))
	}

	lastHeight += govHelper.LazyRewardBlocks()

	// execute block at lastHeight
	req := abcitypes.RequestBeginBlock{
		Header: types2.Header{
			Height: lastHeight,
		},
	}
	bctx := types.NewBlockContext(req, govHelper, acctHelper, nil)
	bctx.AddGas(uint256.NewInt(10))
	_, err := stakeCtrler.EndBlock(bctx)
	require.NoError(t, err)

	_, _, err = stakeCtrler.Commit()
	require.NoError(t, err)

	frozenStakes = stakeCtrler.ReadFrozenStakes()
	require.Equal(t, 0, len(frozenStakes))

	for _, er := range expectedRewards {
		acct1 := acctHelper.FindAccount(er.addr, true)
		require.NotNil(t, acct1)
		require.NotEqual(t, acct1.Balance.Dec(), er.originalBalance.Dec())

		expectedBalance := new(uint256.Int).Add(er.originalBalance, er.rewardedBalance)
		require.Equal(t, expectedBalance.Dec(), acct1.Balance.Dec())

	}
}
