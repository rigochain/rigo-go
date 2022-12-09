package gov

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.MkdirAll(config.DBPath, 0700)

	exitCode := m.Run()

	os.RemoveAll(config.DBPath)

	os.Exit(exitCode)
}
