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
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
	"math"
	"math/big"
	"strconv"
	"sync"
)

var (
	lastBlockHeightKey = []byte("lbh")
)

func blockKey(h int64) []byte {
	return []byte(fmt.Sprintf("bn%v", h))
}

type EVMCtrler struct {
	stateDBWrapper *StateDBWrapper
	ethChainConfig *params.ChainConfig

	metadb          tmdb.DB
	lastRootHash    []byte
	lastBlockHeight int64

	logger tmlog.Logger
	mtx    sync.RWMutex
}

func NewEVMCtrler(path string, acctHandler ctrlertypes.IAccountHandler, logger tmlog.Logger) *EVMCtrler {
	metadb, err := tmdb.NewDB("heightRootHash", "goleveldb", path)
	if err != nil {
		panic(err)
	}
	val, err := metadb.Get(lastBlockHeightKey)
	if err != nil {
		panic(err)
	}
	bn := int64(0)
	if val != nil {
		bn, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			panic(err)
		}
	}
	hash, err := metadb.Get(blockKey(bn))
	if err != nil {
		panic(err)
	}

	logger = logger.With("module", "rigo_EVMCtrler")

	stdb, err := NewStateDBWrapper(path, hash, acctHandler, logger)
	if err != nil {
		panic(err)
	}
	return &EVMCtrler{
		stateDBWrapper:  stdb,
		ethChainConfig:  params.TestChainConfig,
		metadb:          metadb,
		lastRootHash:    hash,
		lastBlockHeight: bn,
		logger:          logger,
	}
}

func (ctrler *EVMCtrler) SetStateDB(state *StateDBWrapper) {
	ctrler.mtx.Lock()
	defer ctrler.mtx.Unlock()
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
	ctrler.lastBlockHeight++
	ctrler.lastRootHash = rootHash[:]

	batch := ctrler.metadb.NewBatch()
	batch.Set(lastBlockHeightKey, []byte(strconv.FormatInt(ctrler.lastBlockHeight, 10)))
	batch.Set(blockKey(ctrler.lastBlockHeight), ctrler.lastRootHash)
	batch.WriteSync()
	batch.Close()

	return rootHash[:], ctrler.lastBlockHeight, nil
}

func (ctrler *EVMCtrler) Close() xerrors.XError {
	if ctrler.metadb != nil {
		if err := ctrler.metadb.Close(); err != nil {
			return xerrors.From(err)
		}
		ctrler.metadb = nil
	}
	return xerrors.From(ctrler.stateDBWrapper.Close())
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
	if ctx.Tx.GetType() != ctrlertypes.TRX_CONTRACT {
		return xerrors.ErrUnknownTrxType
	}

	if ctx.Exec == false {
		ctx.GasUsed = ctx.Tx.Gas
		ctrler.logger.Debug("EVMCtrler is not execute a tx at CheckTx")
		return nil
	}
	ret, xerr := ctrler.execVM(
		ctx.Tx.From,
		ctx.Tx.To,
		ctx.Tx.Nonce,
		ctx.Tx.Amount,
		ctx.Tx.Payload.(*ctrlertypes.TrxPayloadContract).Data,
		ctx.Height,
		ctx.BlockTime,
	)
	if xerr != nil {
		return xerr
	}
	if ret.Err != nil {
		return xerrors.From(ret.Err)
	}
	ctx.RetData = ret.ReturnData

	// `EVM` handles nonce and amount, but not gas.
	ctx.GasUsed = ctx.Tx.Gas //new(uint256.Int).Add(ctx.GasUsed, uint256.NewInt(ret.UsedGas))
	ctrler.stateDBWrapper.SubBalance(ctx.Tx.From.Array20(), ctx.GasUsed.ToBig())

	return nil
}

func (ctrler *EVMCtrler) execVM(from, to types.Address, nonce uint64, amt *uint256.Int, data []byte, height, blockTime int64) (*core.ExecutionResult, xerrors.XError) {
	var sender common.Address
	var toAddr *common.Address
	copy(sender[:], from)
	if to != nil &&
		!types.IsZeroAddress(to) {
		toAddr = new(common.Address)
		copy(toAddr[:], to)
	}

	vmmsg := evmMessage(sender, toAddr, nonce, big.NewInt(0), amt.ToBig(), data)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)
	vmevm := vm.NewEVM(blockContext, txContext, ctrler.stateDBWrapper, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})

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

		ctrler.logger.Info("Create contract", "address", contractAddr)
	}

	return result, nil
}

var _ ctrlertypes.ILedgerHandler = (*EVMCtrler)(nil)
var _ ctrlertypes.ITrxHandler = (*EVMCtrler)(nil)
