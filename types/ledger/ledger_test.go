package ledger

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

type Item struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

func NewItem(nm string, val int32) *Item {
	return &Item{
		Name:  nm,
		Value: val,
	}
}

func (i *Item) GetKey() []byte {
	return []byte(i.Name)
}

func (i *Item) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Item) Unmarshal(d []byte) error {
	return json.Unmarshal(d, i)
}

func (i *Item) Equal(o *Item) bool {
	return (i.Name == o.Name) && (i.Value == o.Value)
}

var _ ILedgerItem = (*Item)(nil)

func TestDefaultTreeLedger(t *testing.T) {
	ledger, err := DefaultTreeLedger[*Item]("treeLedger1", filepath.Join(os.TempDir(), "test"), 256)
	require.NoError(t, err)

	i0 := NewItem("i0", 0)
	i1 := NewItem("i1", 1)
	i2 := NewItem("i2", 2)
	i3 := NewItem("i3", 3)

	ledger.Add(i0)
	ledger.Add(i1)
	ledger.Add(i2)
	_, _, err = ledger.Commit()
	require.NoError(t, err)

	_i0, err := ledger.Find(i0.GetKey())
	require.NoError(t, err)
	require.NotNil(t, _i0)
	require.Equal(t, i0, _i0)

	_i1, err := ledger.Find(i1.GetKey())
	require.NoError(t, err)
	require.NotNil(t, _i1)
	require.Equal(t, i1, _i1)

	_i3, err := ledger.Find(i3.GetKey())
	require.Error(t, err)
	require.Nil(t, _i3)

}
