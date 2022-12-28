package ledger

import (
	"github.com/kysee/arcanus/types/bytes"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"os"
	"path/filepath"
	"testing"
)

var (
	testLedger IFinalityLedger[*MyItem]
	testItem   *MyItem
)

func resetLedger(t *testing.T) {
	dbDir := filepath.Join(os.TempDir(), "test")

	if testLedger != nil {
		require.NoError(t, testLedger.Close())
		os.RemoveAll(dbDir)
	}

	var err error
	testLedger, err = NewFinalityLedger[*MyItem]("treeLedger1", dbDir, 256, func() *MyItem { return &MyItem{} })
	require.NoError(t, err)

	testItem = NewMyItem(bytes.RandHexString(32), rand.Int32())
}

func TestFinalityLedger_Set(t *testing.T) {
	resetLedger(t)

	// do set only
	require.NoError(t, testLedger.Set(testItem))

	// not committed finally
	item, err := testLedger.Get(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)
}

func TestFinalityLedger_SetCommit(t *testing.T) {
	_, _, err := testLedger.Commit()
	require.NoError(t, err)

	// not committed finally
	item, err := testLedger.Get(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)
}

func TestFinalityLedger_SetFinality(t *testing.T) {

	// do set finality
	require.NoError(t, testLedger.SetFinality(testItem))

	// not committed yet
	item, err := testLedger.Get(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)

	item, err = testLedger.GetFinality(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)
}

func TestFinalityLedger_SetFinalityCommit(t *testing.T) {
	// commit finality items
	_, _, err := testLedger.Commit()
	require.NoError(t, err)

	item, err := testLedger.Get(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)

	item, err = testLedger.GetFinality(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)
}

func TestFinalityLedger_Del(t *testing.T) {
	// not exist
	item, err := testLedger.Del(ToLedgerKey(rand.Bytes(32)))
	require.Error(t, err)
	require.Nil(t, item)

	item, err = testLedger.Del(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)

	// not finally deleted yet
	item, err = testLedger.Get(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)
}

func TestFinalityLedger_DelCommit(t *testing.T) {
	// commit finality items
	_, _, err := testLedger.Commit()
	require.NoError(t, err)

	// not finally deleted yet
	item, err := testLedger.Get(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)
}

func TestFinalityLedger_DelFinality(t *testing.T) {
	item, err := testLedger.DelFinality(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)

	// not finally deleted and committed
	item, err = testLedger.Get(testItem.Key())
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testItem, item)

}

func TestFinalityLedger_DelFinalityCommit(t *testing.T) {
	// commit finality items
	_, _, err := testLedger.Commit()
	require.NoError(t, err)

	// not finally deleted and committed
	item, err := testLedger.Get(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)
}

func TestFinalityLedger_CancelSetFinality(t *testing.T) {
	resetLedger(t)

	// do set and delete
	require.NoError(t, testLedger.SetFinality(testItem))
	err := testLedger.CancelSetFinality(testItem.Key())
	require.NoError(t, err)

	_, _, err = testLedger.Commit()
	require.NoError(t, err)

	// not exists
	item, err := testLedger.Get(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)

	// not exists
	item, err = testLedger.GetFinality(testItem.Key())
	require.Error(t, err)
	require.Nil(t, item)

	require.NoError(t, testLedger.Close())
}
