package test

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err = initNode(); err != nil {
			panic(err)
		}
		if err = runNode(); err != nil {
			panic(err)
		}
		wg.Done()
	}()
	wg.Wait()

	time.Sleep(time.Second * 2)

	prepareTest()

	exitCode := m.Run()

	stopNode()
	os.RemoveAll(testConfig.RootDir)

	os.Exit(exitCode)
}
