package evm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"math"
	"sync"
)

type EVMCtrler struct {
	stateDBWrapper *StateDBWrapper
	ethChainConfig *params.ChainConfig

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewEVMCtrler(path string, acctHelper ctrlertypes.IAccountHelper, logger tmlog.Logger) *EVMCtrler {
	stdb, err := NewStateDBWrapper(path, acctHelper)
	if err != nil {
		panic(err)
	}
	return &EVMCtrler{
		stateDBWrapper: stdb,
		ethChainConfig: params.TestChainConfig,
		logger:         logger,
	}
}

func (ctrler *EVMCtrler) InitLedger(req interface{}) xerrors.XError {
	// Handle `lastRoot` at here
	return nil
}

func (ctrler *EVMCtrler) Commit() ([]byte, int64, xerrors.XError) {
	rootHash, err := ctrler.stateDBWrapper.Commit(true)
	if err != nil {
		return nil, 0, xerrors.From(err)
	}
	if err := ctrler.stateDBWrapper.Database().TrieDB().Commit(rootHash, true, nil); err != nil {
		return nil, 0, xerrors.From(err)
	}

	return rootHash[:], 0, nil
}

func (ctrler *EVMCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	//TODO implement me
	panic("implement me")
}

func (ctrler *EVMCtrler) Close() xerrors.XError {
	return xerrors.From(ctrler.stateDBWrapper.StateDB.Database().TrieDB().DiskDB().Close())
}

func (ctrler *EVMCtrler) ValidateTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	if ctx.Tx.GetType() != ctrlertypes.TRX_CONTRACT {
		return xerrors.ErrUnknownTrxType
	}
	payload, ok := ctx.Tx.Payload.(*ctrlertypes.TrxPayloadContract)
	if !ok {
		return xerrors.ErrInvalidTrxPayloadType
	}
	if payload.Data == nil || len(payload.Data) == 0 {
		return xerrors.ErrInvalidTrxPayloadParams
	}
	return nil
}

func (ctrler *EVMCtrler) ExecuteTrx(ctx *ctrlertypes.TrxContext) xerrors.XError {
	ret, xerr := ctrler.callEVM(
		ctx.Tx.From,
		ctx.Tx.To,
		ctx.Tx.Nonce,
		ctx.Tx.Amount,
		ctx.Height,
		ctx.BlockTime,
		ctx.Tx.Payload.(*ctrlertypes.TrxPayloadContract).Data,
	)
	if xerr != nil {
		return xerr
	}
	if ret.Err != nil {
		return xerrors.From(ret.Err)
	}

	_ = ctx.GasUsed.Add(ctx.GasUsed, uint256.NewInt(ret.UsedGas))
	ctx.RetData = ret.ReturnData
	return nil
}

func (ctrler *EVMCtrler) callEVM(from, to types.Address, nonce uint64, amt *uint256.Int, height, blockTime int64, data []byte) (*core.ExecutionResult, xerrors.XError) {
	var sender common.Address
	var toAddr *common.Address
	copy(sender[:], from)
	if to != nil &&
		!types.IsZeroAddress(to) {
		toAddr = new(common.Address)
		copy(toAddr[:], to)
	}

	vmmsg := evmMessage(sender, toAddr, nonce, amt.ToBig(), data)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)
	vmevm := vm.NewEVM(blockContext, vm.TxContext{}, ctrler.stateDBWrapper, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})
	vmevm.Reset(txContext, ctrler.stateDBWrapper)

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

var _ ctrlertypes.ILedgerHandler = (*EVMCtrler)(nil)
var _ ctrlertypes.ITrxHandler = (*EVMCtrler)(nil)
