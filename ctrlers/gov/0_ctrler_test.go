package gov

import (
	cfg "github.com/rigochain/rigo-go/cmd/config"
	"github.com/rigochain/rigo-go/ctrlers/stake"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

var (
	config      = cfg.DefaultConfig()
	govCtrler   *GovCtrler
	stakeHelper *stakeHandlerMock
	acctHelper  *acctHelperMock
	govRule0    = ctrlertypes.DefaultGovRule()
	govRule1    = ctrlertypes.Test1GovRule()
	govRule2    = ctrlertypes.Test2GovRule()
	govRule3    = ctrlertypes.Test3GovRule()
	govRule4    = ctrlertypes.Test4GovRule()
	govRule5    = ctrlertypes.Test5GovRule()
)

func init() {
	config.DBPath = filepath.Join(os.TempDir(), "gov-ctrler-test")
	os.RemoveAll(config.DBPath)
	os.MkdirAll(config.DBPath, 0700)

	var err error
	if govCtrler, err = NewGovCtrler(config, tmlog.NewNopLogger()); err != nil {
		panic(err)
	}
	govCtrler.GovRule = *govRule0

	rand.Seed(time.Now().UnixNano())

	stakeHelper = &stakeHandlerMock{
		valCnt: 5, // 5 delegatees is only validator.
		delegatees: []*stake.Delegatee{
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
			{Addr: types.RandAddress(), TotalPower: rand.Int63n(1000000)},
		},
	}
	sort.Sort(stake.PowerOrderDelegatees(stakeHelper.delegatees))

	acctHelper = &acctHelperMock{
		acctMap: make(map[ctrlertypes.AcctKey]*ctrlertypes.Account),
	}
}

func TestMain(m *testing.M) {
	os.MkdirAll(config.DBPath, 0700)

	exitCode := m.Run()

	os.RemoveAll(config.DBPath)

	os.Exit(exitCode)
}
