package types_test

import (
	"github.com/rigochain/rigo-go/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestConvertAsset(t *testing.T) {
	r := rand.Uint64()
	sau := types.ToSAU(r)
	require.Equal(t, strconv.FormatUint(r, 10)+"000000000000000000", sau.Dec())

	xco, rem := types.FromSAU(sau)
	require.Equal(t, r, xco)
	require.Equal(t, uint64(0), rem)
}
