package types_test

import (
	types2 "github.com/kysee/arcanus/ctrlers/types"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/bytes"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func TestTrxEncode(t *testing.T) {
	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    types.RandAddress(),
		To:      types.RandAddress(),
		Amount:  bytes.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Type:    types2.TRX_TRANSFER,
	}
	require.Equal(t, types2.TRX_TRANSFER, tx0.GetType())

	bzTx0, err := tx0.Encode()
	require.NoError(t, err)

	tx1 := &types2.Trx{}
	err = tx1.Decode(bzTx0)
	require.NoError(t, err)
	require.Equal(t, types2.TRX_TRANSFER, tx1.GetType())
	require.Equal(t, tx0, tx1)

	bzTx1, err := tx1.Encode()
	require.NoError(t, err)
	require.Equal(t, bzTx0, bzTx1)
}

func BenchmarkTrxEncode(b *testing.B) {
	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    types.RandAddress(),
		To:      types.RandAddress(),
		Amount:  bytes.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Payload: &types2.TrxPayloadAssetTransfer{},
	}
	require.Equal(b, types2.TRX_TRANSFER, tx0.GetType())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tx0.Encode()
		require.NoError(b, err)
	}
}

func BenchmarkTrxDecode(b *testing.B) {
	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    types.RandAddress(),
		To:      types.RandAddress(),
		Amount:  bytes.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Payload: &types2.TrxPayloadAssetTransfer{},
	}
	require.Equal(b, types2.TRX_TRANSFER, tx0.GetType())

	bzTx0, err := tx0.Encode()
	require.NoError(b, err)

	tx1 := &types2.Trx{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tx1.Decode(bzTx0)
		require.NoError(b, err)
	}
}
