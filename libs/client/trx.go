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

func NewTrxStaking(from, to types.Address, gas, amt *big.Int, nonce uint64) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		amt,
		gas,
		&trxs.TrxPayloadStaking{})
}

func NewTrxUnstaking(from, to types.Address, gas *big.Int, nonce uint64, txhash types.HexBytes) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		big.NewInt(0),
		gas,
		&trxs.TrxPayloadUnstaking{TxHash: txhash})
}

func NewTrxProposal(from, to types.Address, gas *big.Int, nonce uint64, msg string, blocks int64, propType int32, options [][]byte) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		big.NewInt(0),
		gas,
		&trxs.TrxPayloadProposal{
			Message:      msg,
			VotingBlocks: blocks,
			ProposalType: propType,
			Options:      options,
		})
}

func NewTrxGovRulesProposal(from, to types.Address, gas *big.Int, nonce uint64, msg string, option []byte) *trxs.Trx {
	return NewTrxProposal(from, to, gas, nonce, msg, 10, types.PROPOSAL_GOVRULES, [][]byte{option})
}

func NewTrxVoting(from, to types.Address, gas *big.Int, nonce uint64, txHash types.HexBytes, choice int32) *trxs.Trx {
	return trxs.NewTrx(
		uint32(1),
		from, to,
		nonce,
		big.NewInt(0),
		gas,
		&trxs.TrxPayloadVoting{
			TxHash: txHash,
			Choice: choice,
		})
}
