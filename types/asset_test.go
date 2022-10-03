package types_test

import (
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types"
	"github.com/stretchr/testify/require"
	"math"
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

func TestVotingPower(t *testing.T) {
	amount := libs.RandBigIntN(types.ToSAU(math.MaxInt64))
	xco, _ := types.FromSAU(amount)
	vp := types.AmountToPower(amount)
	require.Equal(t, xco, vp)
}

func BenchmarkGetVotingPower(b *testing.B) {
	amount := libs.RandBigIntN(types.ToSAU(math.MaxInt64))
	xco, _ := types.FromSAU(amount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vp := types.AmountToPower(amount)
		require.Equal(b, xco, vp)
	}
}
