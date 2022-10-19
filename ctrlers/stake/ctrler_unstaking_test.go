package stake

import (
	"github.com/kysee/arcanus/ctrlers/account"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/libs/client"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"math/big"
	"math/rand"
	"os"
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
}

var (
	testGovRules = &gov.GovRules{
		Version:           0,
		MaxValidatorCnt:   21,
		RewardDelayBlocks: 10,
		AmountPerPower:    big.NewInt(1000),
		RewardPerPower:    big.NewInt(10),
	}

	Wallets       []*TestWallet
	stakingTrxs   []*trxs.Trx
	unstakingTrxs []*trxs.Trx

	dummyGas   = big.NewInt(0)
	dummyNonce = uint64(0)

	stakingHistory []*StakingResult
)

func TestMain(m *testing.M) {
	Wallets = makeTestWallet(10000)

	exitCode := m.Run()

	os.Exit(exitCode)
}

func makeTestWallet(n int) []*TestWallet {
	wallets := make([]*TestWallet, n)
	for i := 0; i < n; i++ {
		w := client.NewWallet([]byte("1"))
		a := account.NewAccount(w.Address())

		wallets[i] = &TestWallet{w, a}
	}
	return wallets
}

func makeStakingTrxContext(owners, delegated []*TestWallet) []*trxs.TrxContext {
	var txsCtxs []*trxs.TrxContext
	for _, from := range owners {
		to := delegated[rand.Intn(len(delegated))]
		power := libs.RandInt63n(1000)
		amt := testGovRules.PowerToAmount(power)
		tx := client.NewTrxStaking(from.W.Address(), to.W.Address(), amt, dummyGas, dummyNonce)
		bz, err := tx.Encode()
		if err != nil {
			panic(err)
		}
		txsCtxs = append(txsCtxs, &trxs.TrxContext{
			Tx:           tx,
			TxHash:       crypto.DefaultHash(bz),
			Height:       1,
			SenderPubKey: from.W.GetPubKey(),
			Sender:       from.A,
			Receiver:     to.A,
			NeedAmt:      nil,
			GasUsed:      nil,
			GovRules:     testGovRules,
			Error:        nil,
		})

		stakingHistory = append(stakingHistory, &StakingResult{
			Owner: from.W.Address(),
			To:    to.W.Address(),
			Amt:   amt,
			Power: power,
		})
	}
	return txsCtxs
}

func makeUnstakingTrxContext(owners, delegated []*TestWallet) []*trxs.TrxContext {
	var txsCtxs []*trxs.TrxContext
	for _, from := range owners {
		to := delegated[rand.Intn(len(delegated))]
		power := libs.RandInt63n(1000)
		amt := testGovRules.PowerToAmount(power)
		tx := client.NewTrxStaking(from.W.Address(), to.W.Address(), amt, dummyGas, dummyNonce)
		bz, err := tx.Encode()
		if err != nil {
			panic(err)
		}

		txsCtxs = append(txsCtxs, &trxs.TrxContext{
			Tx:           tx,
			TxHash:       crypto.DefaultHash(bz),
			Height:       1,
			SenderPubKey: from.W.GetPubKey(),
			Sender:       from.A,
			Receiver:     to.A,
			NeedAmt:      nil,
			GasUsed:      nil,
			GovRules:     testGovRules,
			Error:        nil,
		})

		stakingHistory = append(stakingHistory, &StakingResult{
			Owner: from.W.Address(),
			To:    to.W.Address(),
			Amt:   amt,
			Power: power,
		})
	}
	return txsCtxs
}
