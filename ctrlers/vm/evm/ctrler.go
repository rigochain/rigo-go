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
	lastBlockHeightKey       = []byte("lbh")
	RIGOEVMCtrlerChainConfig = &params.ChainConfig{big.NewInt(220819), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil, nil, false, new(params.EthashConfig), nil}
)

func blockKey(h int64) []byte {
	return []byte(fmt.Sprintf("bn%v", h))
}

type EVMCtrler struct {
	ethChainConfig *params.ChainConfig
	ethDB          ethdb.Database
	stateDBWrapper *StateDBWrapper
	acctLedger     ctrlertypes.IAccountHandler

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

	//
	//stdb, err := NewStateDBWrapper(path, hash, acctHandler, logger)
	//if err != nil {
	//	panic(err)
	//}
	return &EVMCtrler{
		ethChainConfig:  RIGOEVMCtrlerChainConfig,
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
	if ctx.Tx.GetType() != ctrlertypes.TRX_CONTRACT {
		return xerrors.ErrUnknownTrxType
	}

	// issue #48
	if xerr := ctrler.stateDBWrapper.Prepare(ctx.TxHash, ctx.TxIdx, ctx.Tx.From, ctx.Tx.To, ctx.Exec); xerr != nil {
		return xerr
	}
	defer ctrler.stateDBWrapper.ApplyTo()

	snap := ctrler.stateDBWrapper.Snapshot()

	ret, xerr := ctrler.execVM(
		ctx.Tx.From,
		ctx.Tx.To,
		ctx.Tx.Nonce,
		ctx.Tx.Gas,
		ctx.Tx.Amount,
		ctx.Tx.Payload.(*ctrlertypes.TrxPayloadContract).Data,
		ctx.Height,
		ctx.BlockTime,
	)
	if xerr != nil {
		ctrler.stateDBWrapper.RevertToSnapshot(snap)
		return xerr
	}
	if ret.Err != nil {
		return xerrors.From(ret.Err)
	}

	ctx.RetData = ret.ReturnData

	// Gas is already applied in EVM.
	// the `EVM` handles nonce, amount and gas.
	ctx.GasUsed = new(uint256.Int).Add(ctx.GasUsed, uint256.NewInt(ret.UsedGas))

	if !ctx.Exec {
		ctrler.stateDBWrapper.RevertToSnapshot(snap)
	} else {
		logs := ctrler.stateDBWrapper.GetLogs(ctx.TxHash.Array32(), common.Hash{})
		if logs != nil && len(logs) > 0 {
			var attrs []abcitypes.EventAttribute
			for _, l := range logs {
				for i, t := range l.Topics {
					strVal := hex.EncodeToString(t.Bytes())
					attrs = append(attrs, abcitypes.EventAttribute{
						Key:   []byte(fmt.Sprintf("topic.%d", i)),
						Value: []byte(strings.ToUpper(strVal)),
						Index: true,
					})
				}
				if l.Data != nil && len(l.Data) > 0 {
					strVal := hex.EncodeToString(l.Data)
					attrs = append(attrs, abcitypes.EventAttribute{
						Key:   []byte("data"),
						Value: []byte(strVal),
						Index: false,
					})
				}
			}
			ctx.Events = append(ctx.Events, abcitypes.Event{
				Type:       "evm",
				Attributes: attrs,
			})
		}
	}

	return nil
}

func (ctrler *EVMCtrler) execVM(from, to types.Address, nonce uint64, gas, amt *uint256.Int, data []byte, height, blockTime int64) (*core.ExecutionResult, xerrors.XError) {
	var sender common.Address
	var toAddr *common.Address
	copy(sender[:], from)
	if to != nil &&
		!types.IsZeroAddress(to) {
		toAddr = new(common.Address)
		copy(toAddr[:], to)
	}

	vmmsg := evmMessage(sender, toAddr, nonce, gas.Uint64(), amt, data)
	blockContext := evmBlockContext(sender, height, blockTime)

	txContext := core.NewEVMTxContext(vmmsg)

	vmevm := vm.NewEVM(blockContext, txContext, ctrler.stateDBWrapper, ctrler.ethChainConfig, vm.Config{NoBaseFee: true})

	gp := new(core.GasPool).AddGas(vmmsg.Gas())
	result, err := core.ApplyMessage(vmevm, vmmsg, gp)
	if err != nil {
		return nil, xerrors.From(err)
	}

	if vmmsg.To() == nil {
		contractAddr := crypto.CreateAddress(vmevm.TxContext.Origin, vmmsg.Nonce())
		result.ReturnData = contractAddr[:]

		ctrler.logger.Debug("Create contract", "address", contractAddr)
	}

	return result, nil
}

func (ctrler *EVMCtrler) EndBlock(context *ctrlertypes.BlockContext) ([]abcitypes.Event, xerrors.XError) {
	return nil, nil
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
