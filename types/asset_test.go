package types_test

import (
	"github.com/rigochain/rigo-go/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestConvertAsset(t *testing.T) {
	r := rand.Int63()
	sau := types.ToSAU(r)
	require.Equal(t, strconv.FormatInt(r, 10)+"000000000000000000", sau.String())

	xco, rem := types.FromSAU(sau)
	require.Equal(t, r, xco)
	require.Equal(t, int64(0), rem)
}
