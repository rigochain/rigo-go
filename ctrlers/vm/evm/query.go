package evm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmrpccore "github.com/tendermint/tendermint/rpc/core"
)

func (ctrler *EVMCtrler) Query(req abcitypes.RequestQuery) ([]byte, xerrors.XError) {
	from := req.Data[:types.AddrSize]
	to := req.Data[types.AddrSize : types.AddrSize*2]
	data := req.Data[types.AddrSize*2:]
	height := req.Height

	if height <= 0 {
		height = ctrler.lastBlockHeight
	}

	block, err := tmrpccore.Block(nil, &height)
	if err != nil {
		return nil, xerrors.From(err)
	}
	btm := block.Block.Time.UnixNano()

	execRet, xerr := ctrler.queryVM(from, to, data, height, btm)
	if xerr != nil {
		return nil, xerr
	}
	if execRet.Err != nil {
		return execRet.Revert(), xerrors.From(execRet.Err)
	}
	//fmt.Printf("return: %x", execRet.Return())
	returnData := &ctrlertypes.VMCallResult{
		execRet.UsedGas,
		execRet.Err,
		execRet.ReturnData,
	}

	retbz, err := tmjson.Marshal(returnData)
	if err != nil {
		return nil, xerrors.From(err)
	}
	return retbz, nil
}

func (ctrler *EVMCtrler) queryVM(from, to types.Address, data []byte, height, blockTime int64) (*core.ExecutionResult, xerrors.XError) {

	// block<height> 시점의 stateDB 와 account ledger(acctCtrler) 를 갖는 `stateDBWrapper` 획득
	hash, err := ctrler.metadb.Get(blockKey(height))
	if err != nil {
		return nil, xerrors.From(err)
	}

	state, xerr := ctrler.ImmutableStateAt(height, hash)
	if xerr != nil {
		return nil, xerr
	}
	xerr = state.Prepare(nil, 0, from, to, false)
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
	vmmsg := evmMessage(sender, toAddr, nonce, uint64(0), uint256.NewInt(0), data)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)
	vmevm := vm.NewEVM(blockContext, txContext, state, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})

	gp := new(core.GasPool).AddGas(vmmsg.Gas())
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
		// contract 생성.
		// EVM은 ReturnData 에 deployed code 를 리턴한다.
		// deployed code 를 contract 주소로 대치 한다.
		contractAddr := crypto.CreateAddress(vmevm.TxContext.Origin, vmmsg.Nonce())
		result.ReturnData = contractAddr[:]
	}

	return result, nil
}
