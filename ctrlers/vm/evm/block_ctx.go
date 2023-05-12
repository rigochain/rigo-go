package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
)

var (
	gasLimit = uint64(1_000_000_000)
)

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	//fmt.Println("========= DEBUG EVM:", "CanTransfer", "address", addr, "amount", amount)
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	//fmt.Println("========= DEBUG EVM:", "Transfer", "sender", sender, "receiver", recipient, "amount", amount)
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

func GetHash(h uint64) common.Hash {
	return common.Hash{}
}

func evmBlockContext(sender common.Address, bn int64, tm int64) vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHash,
		Coinbase:    sender,
		BlockNumber: big.NewInt(bn),
		Time:        big.NewInt(tm),
		Difficulty:  big.NewInt(1),
		BaseFee:     big.NewInt(0),
		GasLimit:    gasLimit * 10_000, //10_000_000_000_000_000_000,
	}
}

func evmMessage(_from common.Address, _to *common.Address, nonce uint64, gasPrice, amt *big.Int, data []byte) types.Message {
	return types.NewMessage(
		_from,
		_to,
		nonce,
		amt,
		gasLimit,      // gas limit
		gasPrice,      // gas price
		big.NewInt(0), // fee cap
		big.NewInt(0), // tip
		data,
		nil,
		false,
	)
}
