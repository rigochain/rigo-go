package types

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProtoCodec(t *testing.T) {
	rule0 := DefaultGovRule()
	bz, err := rule0.Encode()
	require.NoError(t, err)

	rule1, err := DecodeGovRule(bz)
	require.NoError(t, err)

	require.Equal(t, rule0, rule1)

}

func TestJsonCodec(t *testing.T) {
	rule0 := DefaultGovRule()
	bz, err := json.Marshal(rule0)
	require.NoError(t, err)

	rule1 := &GovRule{}
	err = json.Unmarshal(bz, rule1)
	require.NoError(t, err)

	require.Equal(t, rule0, rule1)
}
