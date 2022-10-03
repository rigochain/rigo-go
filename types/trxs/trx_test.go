package trxs_test

import (
	"github.com/kysee/arcanus/libs"
	"github.com/kysee/arcanus/types/trxs"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func TestTrxEncode(t *testing.T) {
	tx0 := &trxs.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    libs.RandAddress(),
		To:      libs.RandAddress(),
		Amount:  libs.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Type:    trxs.TRX_TRANSFER,
	}
	require.Equal(t, trxs.TRX_TRANSFER, tx0.GetType())

	bzTx0, err := tx0.Encode()
	require.NoError(t, err)

	tx1 := &trxs.Trx{}
	err = tx1.Decode(bzTx0)
	require.NoError(t, err)
	require.Equal(t, trxs.TRX_TRANSFER, tx1.GetType())
	require.Equal(t, tx0, tx1)

	bzTx1, err := tx1.Encode()
	require.NoError(t, err)
	require.Equal(t, bzTx0, bzTx1)
}

func BenchmarkTrxEncode(b *testing.B) {
	tx0 := &trxs.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    libs.RandAddress(),
		To:      libs.RandAddress(),
		Amount:  libs.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Payload: &trxs.TrxPayloadAssetTransfer{},
	}
	require.Equal(b, trxs.TRX_TRANSFER, tx0.GetType())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tx0.Encode()
		require.NoError(b, err)
	}
}

func BenchmarkTrxDecode(b *testing.B) {
	tx0 := &trxs.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    libs.RandAddress(),
		To:      libs.RandAddress(),
		Amount:  libs.RandBigInt(),
		Gas:     big.NewInt(rand.Int63()),
		Payload: &trxs.TrxPayloadAssetTransfer{},
	}
	require.Equal(b, trxs.TRX_TRANSFER, tx0.GetType())

	bzTx0, err := tx0.Encode()
	require.NoError(b, err)

	tx1 := &trxs.Trx{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = tx1.Decode(bzTx0)
		require.NoError(b, err)
	}
}
