package gov

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCodec(t *testing.T) {
	rule0 := DefaultGovRules()
	bz, err := rule0.Encode()
	require.NoError(t, err)

	rule1, err := DecodeGovRules(bz)
	require.NoError(t, err)

	require.Equal(t, rule0, rule1)

}
