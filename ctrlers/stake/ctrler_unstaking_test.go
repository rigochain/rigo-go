package stake

import (
	"bytes"
	"errors"
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

type TestWallet struct {
	W *client.Wallet
	A types.IAccount
}

type StakingResult struct {
	Owner  types.Address
	To     types.Address
	Amt    *big.Int
	Power  int64
	TxHash types.HexBytes
	Height int64
}

var (
	DBDIR          = filepath.Join(os.TempDir(), "stake-ctrler-test")
	stakeCtrler, _ = NewStakeCtrler(DBDIR, log.NewNopLogger())
	testGovRules   = &gov.GovRules{
		Version:           0,
		MaxValidatorCnt:   21,
		RewardDelayBlocks: 10,
		AmountPerPower:    big.NewInt(1000),
		RewardPerPower:    big.NewInt(10),
	}

	Wallets          []*TestWallet
	stakingTrxCtxs   []*trxs.TrxContext
	unstakingTrxCtxs []*trxs.TrxContext

	dummyGas   = big.NewInt(0)
	dummyNonce = uint64(0)

	stakingHistory []*StakingResult
	lastHeight     = int64(1)
)

func TestMain(m *testing.M) {

	Wallets = makeTestWallet(10000)

	for i := 0; i < 1000; i++ {
		if txctx, err := randMakeStakingTrxContext(); err != nil {
			panic(err)
		} else {
			stakingTrxCtxs = append(stakingTrxCtxs, txctx)
		}
		if rand.Intn(10000000)%3 == 0 {
			lastHeight++
		}
	}
	lastHeight += 10
	for i := 0; i < 100; i++ {
		if txctx, err := randMakeUnstakingTrxContext(); err != nil {
			panic(err)
		} else {
			unstakingTrxCtxs = append(unstakingTrxCtxs, txctx)
		}
		if rand.Intn(10000000)%3 == 0 {
			lastHeight++
		}
	}

	os.RemoveAll(DBDIR)

	exitCode := m.Run()

	os.RemoveAll(DBDIR)

	os.Exit(exitCode)
}

func TestStaking(t *testing.T) {
	sumAmt := big.NewInt(0)
	sumPower := int64(0)
	for _, txctx := range stakingTrxCtxs {
		err := stakeCtrler.Apply(txctx)
		require.NoError(t, err)

		sumAmt.Add(sumAmt, txctx.Tx.Amount)
		sumPower += txctx.GovRules.AmountToPower(txctx.Tx.Amount)
	}

	require.Equal(t, sumAmt.String(), stakeCtrler.GetTotalAmount().String())
	require.Equal(t, sumPower, stakeCtrler.GetTotalPower())
}

func makeTestWallet(n int) []*TestWallet {
	wallets := make([]*TestWallet, n)
	for i := 0; i < n; i++ {
		w := client.NewWallet([]byte("1"))
		a := account.NewAccount(w.Address())
		a.AddBalance(types.ToSAU(100000000))

		wallets[i] = &TestWallet{w, a}
	}
	return wallets
}

func findTestWallet(addr types.Address) *TestWallet {
	for _, w := range Wallets {
		if bytes.Compare(addr, w.W.Address()) == 0 {
			return w
		}
	}
	return nil
}

func randMakeStakingTrxContext() (*trxs.TrxContext, error) {
	rn0 := rand.Intn(len(Wallets))
	rn1 := rand.Intn(len(Wallets))
	if rn0 == rn1 {
		rn1 = (rn1 + 1) % len(Wallets)
	}

	from := Wallets[rn0]
	to := Wallets[rn1]

	power := libs.RandInt63n(1000)

	return makeStakingTrxContext(from, to, power)
}

func makeStakingTrxContext(from, to *TestWallet, power int64) (*trxs.TrxContext, error) {
	amt := testGovRules.PowerToAmount(power)

	tx := client.NewTrxStaking(from.W.Address(), to.W.Address(), amt, dummyGas, dummyNonce)
	bz, err := tx.Encode()
	if err != nil {
		return nil, err
	}

	stakingHistory = append(stakingHistory, &StakingResult{
		Owner: from.W.Address(),
		To:    to.W.Address(),
		Amt:   amt,
		Power: power,
	})

	return &trxs.TrxContext{
		Tx:           tx,
		TxHash:       crypto.DefaultHash(bz),
		Height:       lastHeight,
		SenderPubKey: from.W.GetPubKey(),
		Sender:       from.A,
		Receiver:     to.A,
		NeedAmt:      nil,
		GasUsed:      nil,
		GovRules:     testGovRules,
		Error:        nil,
	}, nil
}

func randMakeUnstakingTrxContext() (*trxs.TrxContext, error) {
	rn := rand.Intn(len(stakingHistory))
	unstakingStake := stakingHistory[rn]
	from := findTestWallet(unstakingStake.Owner)
	if from == nil {
		return nil, errors.New("not found test account for " + unstakingStake.Owner.String())
	}
	to := findTestWallet(unstakingStake.To)
	if to == nil {
		return nil, errors.New("not found test account for " + unstakingStake.To.String())
	}

	return makeUnstakingTrxContext(from, to, unstakingStake.TxHash)
}

func makeUnstakingTrxContext(from, to *TestWallet, txhash types.HexBytes) (*trxs.TrxContext, error) {

	tx := client.NewTrxUnstaking(from.W.Address(), to.W.Address(), txhash, dummyGas, dummyNonce)
	tzbz, _, err := from.W.SignTrx(tx)
	if err != nil {
		return nil, err
	}

	return &trxs.TrxContext{
		Tx:           tx,
		TxHash:       crypto.DefaultHash(tzbz),
		Height:       lastHeight,
		SenderPubKey: from.W.GetPubKey(),
		Sender:       from.A,
		Receiver:     to.A,
		GovRules:     testGovRules,
	}, nil
}
