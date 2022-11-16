package gov

import (
	tmlog "github.com/tendermint/tendermint/libs/log"
	"os"
	"path/filepath"
	"testing"
)

var (
	DBDIR     = filepath.Join(os.TempDir(), "gov-ctrler-test")
	govCtrler *GovCtrler
)

func TestMain(m *testing.M) {

	var err error
	if govCtrler, err = NewGovCtrler(DBDIR, tmlog.NewNopLogger()); err != nil {
		panic(err)
	}
	govCtrler.SetRules(DefaultGovRule())

	exitCode := m.Run()

	os.RemoveAll(DBDIR)

	os.Exit(exitCode)
}
