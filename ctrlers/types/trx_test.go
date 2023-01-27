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
	//    js: 08011080d1b5f9b49ce09e17180122141852deff9121030a774055ec9ed344f19c19c1902a141852deff9121030a774055ec9ed344f19c19c1903201003a010a40034a220a20e4700ca2c8a0889341b66f5962a8949ed079b41420ee43575944c88504088505
	//  node: 080110C0E0F9FB849A859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD
	//client: 080110809E8DA1B88F859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713201003A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD
	bz0, err := hex.DecodeString("08011080bfb7f6e6c8e09e17180122142f3ceb3253f8ae16d2082d85556aed242575471f2a142f3ceb3253f8ae16d2082d85556aed242575471f3201003a010a40034a220a208e59ec501496d44e1cd8508d8ed62afd2013d30dc4e97def9e228a02c6e6277f5241ce47b89d5d34a5c27cca2e1aa9384f531cfae7a2341050399d9d6bd6f9d477f41b5f1791af34f0c9ca37eca8dcd17dd84df425f47667a7a37cdf07d02b48547600")
	require.NoError(t, err)
	encoded0 := bytes.HexBytes(bz0)

	//bz1, err := hex.DecodeString("080110809E8DA1B88F859D1718032214E8CF086D07A8A742556E356C54641495B57118712A14E8CF086D07A8A742556E356C54641495B57118713201003A010A40034A220A20506722A75DE1062B937A351025C278327411392080521DC99BD86592DD3151BD")
	//require.NoError(t, err)
	//encoded1 := bytes.HexBytes(bz1)

	decoded0 := &types2.Trx{}
	require.NoError(t, decoded0.Decode(encoded0))

	//decoded1 := &types2.Trx{}
	//require.NoError(t, decoded1.Decode(encoded1))

	encoded2, err := decoded0.Encode()
	require.NoError(t, err)
	require.Equal(t, encoded0, bytes.HexBytes(encoded2))

	//decoded2 := &types2.Trx{}
	//require.NoError(t, decoded2.Decode(encoded2))
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
