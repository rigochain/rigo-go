package types_test

import (
	"encoding/hex"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
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

func TestTrxPayloadUnstaking_Decode(t *testing.T) {
	hexFromJS := "080110c0cfb29bcbb08ba01718012214ef6f2d241e32243c84dbc937a56031444498f8e22a14ef6f2d241e32243c84dbc937a56031444498f8e23a010a40034a220a207b511d1c04be7e8002f20784e34f39814c8179578497953f8402320a16cea3df"
	bz0, err := hex.DecodeString(hexFromJS)
	require.NoError(t, err)
	encoded0 := bytes.HexBytes(bz0)

	decoded0 := &types2.Trx{}
	require.NoError(t, decoded0.Decode(encoded0))

	encoded2, err := decoded0.Encode()
	require.NoError(t, err)
	require.Equal(t, encoded0, bytes.HexBytes(encoded2))
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
