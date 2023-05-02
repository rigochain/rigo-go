package web3

import (
	"github.com/holiman/uint256"
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
)

func NewTrxTransfer(from, to types.Address, nonce uint64, gas, amt *uint256.Int) *types2.Trx {
	return types2.NewTrx(
		uint32(1),
		from, to,
		nonce,
		gas,
		amt,
		&types2.TrxPayloadAssetTransfer{})
}

func NewTrxStaking(from, to types.Address, nonce uint64, gas, amt *uint256.Int) *types2.Trx {
	return types2.NewTrx(
		uint32(1),
		from, to,
		nonce,
		gas,
		amt,
		&types2.TrxPayloadStaking{})
}

func NewTrxUnstaking(from, to types.Address, nonce uint64, gas *uint256.Int, txhash bytes.HexBytes) *types2.Trx {
	return types2.NewTrx(
		uint32(1),
		from, to,
		nonce,
		gas,
		uint256.NewInt(0),
		&types2.TrxPayloadUnstaking{TxHash: txhash})
}

func NewTrxProposal(from, to types.Address, nonce uint64, gas *uint256.Int, msg string, start, period int64, optType int32, options ...[]byte) *types2.Trx {
	return types2.NewTrx(
		uint32(1),
		from, to,
		nonce,
		gas,
		uint256.NewInt(0),
		&types2.TrxPayloadProposal{
			Message:            msg,
			StartVotingHeight:  start,
			VotingPeriodBlocks: period,
			OptType:            optType,
			Options:            options,
		})
}

func NewTrxVoting(from, to types.Address, nonce uint64, gas *uint256.Int, txHash bytes.HexBytes, choice int32) *types2.Trx {
	return types2.NewTrx(
		uint32(1),
		from, to,
		nonce,
		gas,
		uint256.NewInt(0),
		&types2.TrxPayloadVoting{
			TxHash: txHash,
			Choice: choice,
		})
}
