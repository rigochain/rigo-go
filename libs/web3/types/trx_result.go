package types

import (
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type TrxResult struct {
	*coretypes.ResultTx
	TrxObj *ctrlertypes.Trx `json:"trx_obj"`
}
