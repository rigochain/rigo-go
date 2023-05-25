package types_test

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/libs/web3"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/stretchr/testify/require"
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
		Amount:  bytes.RandU256Int(),
		Gas:     uint256.NewInt(rand.Uint64()),
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

func TestRLP_TrxPayloadContract(t *testing.T) {
	w := web3.NewWallet([]byte("1"))
	require.NoError(t, w.Unlock([]byte("1")))

	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    w.Address(),
		To:      types.RandAddress(),
		Amount:  bytes.RandU256Int(),
		Gas:     uint256.NewInt(rand.Uint64()),
		Type:    types2.TRX_CONTRACT,
		Payload: &types2.TrxPayloadContract{
			Data: bytes.RandHexBytes(10234),
		},
	}
	_, _, err := w.SignTrxRLP(tx0)

	require.NoError(t, err)
	require.NoError(t, types2.VerifyTrxRLP(tx0))

	bz0, err := rlp.EncodeToBytes(tx0)
	require.NoError(t, err)

	tx1 := &types2.Trx{}
	err = rlp.DecodeBytes(bz0, tx1)
	require.NoError(t, err)
	require.NoError(t, types2.VerifyTrxRLP(tx1))

	require.Equal(t,
		tx0.Payload.(*types2.TrxPayloadContract).Data,
		tx1.Payload.(*types2.TrxPayloadContract).Data)

	bz1, err := rlp.EncodeToBytes(tx1)
	require.Equal(t, bz0, bz1)
}

func TestRLP_TrxPayloadSetDoc(t *testing.T) {
	w := web3.NewWallet([]byte("1"))
	require.NoError(t, w.Unlock([]byte("1")))

	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    w.Address(),
		To:      types.RandAddress(),
		Amount:  bytes.RandU256Int(),
		Gas:     uint256.NewInt(rand.Uint64()),
		Type:    types2.TRX_SETDOC,
		Payload: &types2.TrxPayloadSetDoc{
			Name: "test account doc",
			URL:  "https://test.account.doc/1",
		},
	}
	_, _, err := w.SignTrxRLP(tx0)

	require.NoError(t, err)
	require.NoError(t, types2.VerifyTrxRLP(tx0))

	bz0, err := rlp.EncodeToBytes(tx0)
	require.NoError(t, err)

	tx1 := &types2.Trx{}
	err = rlp.DecodeBytes(bz0, tx1)
	require.NoError(t, err)
	require.NoError(t, types2.VerifyTrxRLP(tx1))

	require.Equal(t,
		tx0.Payload.(*types2.TrxPayloadSetDoc).Name,
		tx1.Payload.(*types2.TrxPayloadSetDoc).Name)

	require.Equal(t,
		tx0.Payload.(*types2.TrxPayloadSetDoc).URL,
		tx1.Payload.(*types2.TrxPayloadSetDoc).URL)

	bz1, err := rlp.EncodeToBytes(tx1)
	require.Equal(t, bz0, bz1)
}

func BenchmarkTrxEncode(b *testing.B) {
	tx0 := &types2.Trx{
		Version: 1,
		Time:    time.Now().UnixNano(),
		Nonce:   rand.Uint64(),
		From:    types.RandAddress(),
		To:      types.RandAddress(),
		Amount:  bytes.RandU256Int(),
		Gas:     uint256.NewInt(rand.Uint64()),
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
		Amount:  bytes.RandU256Int(),
		Gas:     uint256.NewInt(rand.Uint64()),
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
