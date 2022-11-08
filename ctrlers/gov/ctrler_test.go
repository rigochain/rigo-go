package gov

import (
	tmlog "github.com/tendermint/tendermint/libs/log"
	"os"
	"path/filepath"
	"testing"
)

var (
	DBDIR        = filepath.Join(os.TempDir(), "gov-ctrler-test")
	govCtrler, _ = NewGovCtrler(DBDIR, tmlog.NewNopLogger())
	govRules     = DefaultGovRules()
)

func TestMain(m *testing.M) {
	//os.RemoveAll(DBDIR)

	exitCode := m.Run()

	os.RemoveAll(DBDIR)

	os.Exit(exitCode)
}
