package types

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"github.com/rigochain/rigo-go/types/xerrors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type BlockContext struct {
	BlockInfo   abcitypes.RequestBeginBlock `json:"blockInfo"`
	Fee         *uint256.Int                `json:"fee"`
	TxsCnt      int                         `json:"TxsCnt"`
	GovHelper   IGovHelper
	AcctHelper  IAccountHelper
	StakeHelper IStakeHelper

	ValUpdates abcitypes.ValidatorUpdates
}

func (bctx *BlockContext) MarshalJSON() ([]byte, error) {
	_bctx := &struct {
		BlockInfo abcitypes.RequestBeginBlock `json:"blockInfo"`
		Fee       *uint256.Int                `json:"fee"`
	}{
		BlockInfo: bctx.BlockInfo,
		Fee:       bctx.Fee,
	}

	return json.Marshal(_bctx)
}

func (bctx *BlockContext) UnmarshalJSON(bz []byte) error {
	_bctx := &struct {
		BlockInfo abcitypes.RequestBeginBlock `json:"blockInfo"`
		Fee       *uint256.Int                `json:"fee"`
	}{}

	if err := json.Unmarshal(bz, _bctx); err != nil {
		return err
	}
	bctx.BlockInfo = _bctx.BlockInfo
	bctx.Fee = _bctx.Fee
	return nil
}

type IBlockHandler interface {
	ValidateBlock(*BlockContext) xerrors.XError
	ExecuteBlock(*BlockContext) xerrors.XError
}
