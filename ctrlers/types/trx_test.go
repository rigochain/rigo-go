package types_test

import (
	"encoding/hex"
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

func TestTrxPayloadUnstaking_Decode(t *testing.T) {
	//node:   080110C0E0F9FB849A859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD
	//client: 080110809E8DA1B88F859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713201003A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD
	bz0, err := hex.DecodeString("080110C0E0F9FB849A859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD")
	require.NoError(t, err)
	encoded0 := bytes.HexBytes(bz0)

	bz1, err := hex.DecodeString("080110809E8DA1B88F859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713201003A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD")
	require.NoError(t, err)
	encoded1 := bytes.HexBytes(bz1)

	decoded0 := &types2.Trx{}
	require.NoError(t, decoded0.Decode(encoded0))

	decoded1 := &types2.Trx{}
	require.NoError(t, decoded1.Decode(encoded1))

	bz2, err := decoded1.Encode()
	require.NoError(t, err)

	require.Equal(t, bz1, bz2)

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
