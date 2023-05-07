package types

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"sync"
)

type BlockContext struct {
	blockInfo abcitypes.RequestBeginBlock
	gasSum    *uint256.Int
	txsCnt    int
	appHash   bytes.HexBytes

	GovHelper   IGovHelper
	AcctHelper  IAccountHelper
	StakeHelper IStakeHelper

	ValUpdates abcitypes.ValidatorUpdates

	mtx sync.RWMutex
}

func NewBlockContext(bi abcitypes.RequestBeginBlock, g IGovHelper, a IAccountHelper, s IStakeHelper) *BlockContext {
	return &BlockContext{
		blockInfo:   bi,
		gasSum:      uint256.NewInt(0),
		txsCnt:      0,
		appHash:     nil,
		GovHelper:   g,
		AcctHelper:  a,
		StakeHelper: s,
		ValUpdates:  nil,
	}
}

func (bctx *BlockContext) BlockInfo() abcitypes.RequestBeginBlock {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.blockInfo
}

func (bctx *BlockContext) SetHeight(h int64) {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	bctx.blockInfo.Header.Height = h
}

func (bctx *BlockContext) Height() int64 {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.blockInfo.Header.Height
}

func (bctx *BlockContext) PreAppHash() bytes.HexBytes {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.blockInfo.Header.GetAppHash()
}

func (bctx *BlockContext) AppHash() bytes.HexBytes {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.appHash
}

func (bctx *BlockContext) SetAppHash(hash []byte) {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	bctx.appHash = hash
}

func (bctx *BlockContext) TimeNano() int64 {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.blockInfo.Header.GetTime().UnixNano()
}

func (bctx *BlockContext) GasSum() *uint256.Int {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.gasSum.Clone()
}

func (bctx *BlockContext) AddGas(gas *uint256.Int) {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	_ = bctx.gasSum.Add(bctx.gasSum, gas)
}

func (bctx *BlockContext) TxsCnt() int {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.txsCnt
}

func (bctx *BlockContext) AddTxsCnt(d int) {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	bctx.txsCnt += d
}

func (bctx *BlockContext) GetValUpdates() abcitypes.ValidatorUpdates {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	return bctx.ValUpdates
}

func (bctx *BlockContext) SetValUpdates(valUps abcitypes.ValidatorUpdates) {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	bctx.ValUpdates = valUps
}

func (bctx *BlockContext) MarshalJSON() ([]byte, error) {
	bctx.mtx.RLock()
	defer bctx.mtx.RUnlock()

	_bctx := &struct {
		BlockInfo abcitypes.RequestBeginBlock `json:"blockInfo"`
		GasSum    *uint256.Int                `json:"gasSum"`
		TxsCnt    int                         `json:"txsCnt"`
		AppHash   []byte                      `json:"appHash"`
	}{
		BlockInfo: bctx.blockInfo,
		GasSum:    bctx.gasSum,
		TxsCnt:    bctx.txsCnt,
		AppHash:   bctx.appHash,
	}

	return json.Marshal(_bctx)
}

func (bctx *BlockContext) UnmarshalJSON(bz []byte) error {
	bctx.mtx.Lock()
	defer bctx.mtx.Unlock()

	_bctx := &struct {
		BlockInfo abcitypes.RequestBeginBlock `json:"blockInfo"`
		GasSum    *uint256.Int                `json:"gasSum"`
		TxsCnt    int                         `json:"txsCnt"`
		AppHash   []byte                      `json:"appHash"`
	}{}

	if err := json.Unmarshal(bz, _bctx); err != nil {
		return err
	}
	bctx.blockInfo = _bctx.BlockInfo
	bctx.gasSum = _bctx.GasSum
	bctx.txsCnt = _bctx.TxsCnt
	bctx.appHash = _bctx.AppHash
	return nil
}

type IBlockHandler interface {
	ValidateBlock(*BlockContext) xerrors.XError
	ExecuteBlock(*BlockContext) xerrors.XError
}
