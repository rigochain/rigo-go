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
		trxs.TRX_TRANSFER)
}

func NewTrxStaking(from, to types.Address, amt, gas *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		amt,
		gas,
		trxs.TRX_STAKING)
}

func NewTrxUnstaking(from, to types.Address, amt, gas *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		amt,
		gas,
		trxs.TRX_STAKING)
}
