package evm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"math"
	"math/big"
)

func (ctrler *EVMCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	return nil, nil
}

func (ctrler *EVMCtrler) queryVM(from, to types.Address, data []byte, height, blockTime int64) (*core.ExecutionResult, xerrors.XError) {

	// block<height> 시점의 stateDB 와 account ledger(acctCtrler) 를 갖는 `stateDBWrapper` 획득
	hash, err := ctrler.metadb.Get(blockKey(height))
	if err != nil {
		return nil, xerrors.From(err)
	}

	state, xerr := ctrler.stateDBWrapper.ImmutableStateAt(height, hash)
	if xerr != nil {
		return nil, xerr
	}
	defer func() { state = nil }()

	var sender common.Address
	var toAddr *common.Address
	copy(sender[:], from)
	if to != nil &&
		!types.IsZeroAddress(to) {
		toAddr = new(common.Address)
		copy(toAddr[:], to)
	}

	nonce := state.GetNonce(from.Array20())

	vmmsg := evmMessage(sender, toAddr, nonce, big.NewInt(0), data)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)
	vmevm := vm.NewEVM(blockContext, txContext, state, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})

	gp := new(core.GasPool).AddGas(math.MaxUint64)
	result, err := core.ApplyMessage(vmevm, vmmsg, gp)
	if err != nil {
		return nil, xerrors.From(err)
	}

	// If the timer caused an abort, return an appropriate error message
	if vmevm.Cancelled() {
		return nil, xerrors.From(fmt.Errorf("execution aborted (timeout ???)"))
	}
	if err != nil {
		return nil, xerrors.From(fmt.Errorf("err: %w (supplied gas %d)", err, vmmsg.Gas()))
	}

	if vmmsg.To() == nil {
		contractAddr := crypto.CreateAddress(vmevm.TxContext.Origin, vmmsg.Nonce())
		result.ReturnData = contractAddr[:]
	}

	return result, nil
}
