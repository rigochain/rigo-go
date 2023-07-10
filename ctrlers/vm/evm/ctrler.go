package evm

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

var (
	lastBlockHeightKey              = []byte("lbh")
	RIGOTestnetEVMCtrlerChainConfig = &params.ChainConfig{big.NewInt(220818), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil, nil, false, new(params.EthashConfig), nil}
	RIGOMainnetEVMCtrlerChainConfig = &params.ChainConfig{big.NewInt(220819), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil, nil, false, new(params.EthashConfig), nil}
)

func blockKey(h int64) []byte {
	return []byte(fmt.Sprintf("bn%v", h))
}

type EVMCtrler struct {
	ethChainConfig *params.ChainConfig
	ethDB          ethdb.Database
	stateDBWrapper *StateDBWrapper
	acctLedger     ctrlertypes.IAccountHandler
	blockGasPool   *core.GasPool

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

	//rawDB, err := rawdb.NewLevelDBDatabaseWithFreezer(path, 0, 0, path, "", false)
	db, err := rawdb.NewLevelDBDatabase(path, 128, 128, "", false)
	if err != nil {
		panic(err)
	}

	logger = logger.With("module", "rigo_EVMCtrler")

	return &EVMCtrler{
		ethChainConfig:  RIGOMainnetEVMCtrlerChainConfig,
		ethDB:           db,
		metadb:          metadb,
		acctLedger:      acctHandler,
		lastRootHash:    hash,
		lastBlockHeight: bn,
		logger:          logger,
	}
}

func (ctrler *EVMCtrler) InitLedger(req interface{}) xerrors.XError {
	// Handle `lastRoot` at here
	return nil
}

func (ctrler *EVMCtrler) BeginBlock(ctx *ctrlertypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	if ctrler.lastBlockHeight+1 != ctx.Height() {
		return nil, xerrors.ErrBeginBlock.Wrapf("wrong block height - expected: %v, actual: %v", ctrler.lastBlockHeight+1, ctx.Height())
	}

	stdb, err := NewStateDBWrapper(ctrler.ethDB, ctrler.lastRootHash, ctx.AcctHandler, ctrler.logger)
	if err != nil {
		return nil, xerrors.From(err)
	}

	ctrler.stateDBWrapper = stdb
	ctrler.blockGasPool = new(core.GasPool).AddGas(gasLimit)
	return nil, nil
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
	if ctx.Exec == false {
		// issue #71
		// Execute a contract transaction only on `deliveryTx`
		return nil
	}
	if ctx.Tx.GetType() != ctrlertypes.TRX_CONTRACT {
		return xerrors.ErrUnknownTrxType
	}

	// issue #69 - in order to pass `snap` to `Prepare`, call `Snapshot` before `Prepare`
	snap := ctrler.stateDBWrapper.Snapshot()
	// issue #48 - prepare hash and index of tx
	ctrler.stateDBWrapper.Prepare(ctx.TxHash, ctx.TxIdx, ctx.Tx.From, ctx.Tx.To, snap, ctx.Exec)

	ret, xerr := ctrler.execVM(
		ctx.Tx.From,
		ctx.Tx.To,
		ctx.Tx.Nonce,
		feeToGas(ctx.Tx.Gas, ctx.GovHandler.GasPrice()),
		ctx.GovHandler.GasPrice(),
		ctx.Tx.Amount,
		ctx.Tx.Payload.(*ctrlertypes.TrxPayloadContract).Data,
		ctx.Height,
		ctx.BlockTime,
		ctx.Exec,
	)
	if xerr != nil {
		ctrler.stateDBWrapper.RevertToSnapshot(snap)
		ctrler.stateDBWrapper.Finish()
		return xerr
	}

	if ret.Failed() {
		ctrler.stateDBWrapper.RevertToSnapshot(snap)
		ctrler.stateDBWrapper.Finish()
		return xerrors.From(ret.Err)
	}

	ctrler.stateDBWrapper.Finish()

	// Update the state with pending changes.
	blockNumber := uint256.NewInt(uint64(ctx.Height)).ToBig()
	if ctrler.ethChainConfig.IsByzantium(blockNumber) {
		ctrler.stateDBWrapper.Finalise(true)
	} else {
		ctrler.lastRootHash = ctrler.stateDBWrapper.IntermediateRoot(ctrler.ethChainConfig.IsEIP158(blockNumber)).Bytes()
	}

	ctx.RetData = ret.ReturnData

	// Gas is already applied by buyGas and refundGas of EVM.
	// the `EVM` handles nonce, amount and gas.
	ctx.GasUsed = new(uint256.Int).Add(ctx.GasUsed, gasToFee(ret.UsedGas, ctx.GovHandler.GasPrice()))

	// Add events from evm.
	if ctx.Exec {
		logs := ctrler.stateDBWrapper.GetLogs(ctx.TxHash.Array32(), common.Hash{})
		if logs != nil && len(logs) > 0 {
			var attrs []abcitypes.EventAttribute
			for _, l := range logs {
				// Contract Address
				strVal := hex.EncodeToString(l.Address[:])
				attrs = append(attrs, abcitypes.EventAttribute{
					Key:   []byte("contract"),
					Value: []byte(strVal),
					Index: true,
				})

				// Topics (indexed)
				for i, t := range l.Topics {
					strVal := hex.EncodeToString(t.Bytes())
					attrs = append(attrs, abcitypes.EventAttribute{
						Key:   []byte(fmt.Sprintf("topic.%d", i)),
						Value: []byte(strings.ToUpper(strVal)),
						Index: true,
					})
				}

				// Data (not indexed)
				if l.Data != nil && len(l.Data) > 0 {
					strVal := hex.EncodeToString(l.Data)
					attrs = append(attrs, abcitypes.EventAttribute{
						Key:   []byte("data"),
						Value: []byte(strVal),
						Index: false,
					})
				}

				// Removed
				strVal = "false"
				if l.Removed {
					strVal = "true"
				}
				attrs = append(attrs, abcitypes.EventAttribute{
					Key:   []byte("removed"),
					Value: []byte(strVal),
					Index: false,
				})
			}
			ctx.Events = append(ctx.Events, abcitypes.Event{
				Type:       "evm",
				Attributes: attrs,
			})
		}
	}

	return nil
}

func (ctrler *EVMCtrler) execVM(from, to types.Address, nonce, gas uint64, gasPrice, amt *uint256.Int, data []byte, height, blockTime int64, exec bool) (*core.ExecutionResult, xerrors.XError) {
	var sender common.Address
	var toAddr *common.Address
	copy(sender[:], from)
	if to != nil &&
		!types.IsZeroAddress(to) {
		toAddr = new(common.Address)
		copy(toAddr[:], to)
	}

	vmmsg := evmMessage(sender, toAddr, nonce, gas, gasPrice, amt, data, false)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)

	vmevm := vm.NewEVM(blockContext, txContext, ctrler.stateDBWrapper, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})

	result, err := core.ApplyMessage(vmevm, vmmsg, ctrler.blockGasPool)
	if err != nil {
		return nil, xerrors.From(err)
	}

	if vmmsg.To() == nil && !result.Failed() {
		contractAddr := crypto.CreateAddress(vmevm.TxContext.Origin, vmmsg.Nonce())
		result.ReturnData = contractAddr[:]
		ctrler.logger.Debug("Create contract", "address", contractAddr)
	}

	return result, nil
}

func (ctrler *EVMCtrler) EndBlock(context *ctrlertypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	ctrler.blockGasPool = new(core.GasPool).AddGas(gasLimit)
	return nil, nil
}

func (ctrler *EVMCtrler) Commit() ([]byte, int64, xerrors.XError) {
	rootHash, err := ctrler.stateDBWrapper.Commit(true)
	if err != nil {
		panic(err)
	}
	if err := ctrler.stateDBWrapper.Database().TrieDB().Commit(rootHash, true, nil); err != nil {
		panic(err)
	}
	ctrler.lastBlockHeight++
	ctrler.lastRootHash = rootHash[:]

	batch := ctrler.metadb.NewBatch()
	batch.Set(lastBlockHeightKey, []byte(strconv.FormatInt(ctrler.lastBlockHeight, 10)))
	batch.Set(blockKey(ctrler.lastBlockHeight), ctrler.lastRootHash)
	batch.WriteSync()
	batch.Close()

	stdb, err := NewStateDBWrapper(ctrler.ethDB, ctrler.lastRootHash, ctrler.acctLedger, ctrler.logger)
	if err != nil {
		panic(err)
	}

	ctrler.stateDBWrapper = stdb

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

func (ctrler *EVMCtrler) ImmutableStateAt(n int64, hash []byte) (*StateDBWrapper, xerrors.XError) {
	rootHash := bytes.HexBytes(hash).Array32()
	stateDB, err := state.New(rootHash, state.NewDatabase(ctrler.ethDB), nil)
	if err != nil {
		return nil, xerrors.From(err)
	}

	acctLedger, xerr := ctrler.acctLedger.ImmutableAcctCtrlerAt(n)
	if xerr != nil {
		return nil, xerr
	}
	return &StateDBWrapper{
		StateDB:          stateDB,
		acctLedger:       acctLedger,
		accessedObjAddrs: make(map[common.Address]int),
		immutable:        true,
		logger:           ctrler.logger,
	}, nil
}

var _ ctrlertypes.ILedgerHandler = (*EVMCtrler)(nil)
var _ ctrlertypes.ITrxHandler = (*EVMCtrler)(nil)
var _ ctrlertypes.IBlockHandler = (*EVMCtrler)(nil)
