package client

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/trxs"
	"math/big"
)

func NewTrxTransfer(from, to types.Address, amt, gas *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		amt,
		gas,
		&trxs.TrxPayloadAssetTransfer{})
}

func NewTrxStaking(from, to types.Address, amt, gas *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		amt,
		gas,
		&trxs.TrxPayloadStaking{})
}

func NewTrxUnstaking(from, to types.Address, txhash types.HexBytes, gas *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		big.NewInt(0),
		gas,
		&trxs.TrxPayloadUnstaking{TxHash: txhash})
}
