package types

import (
	ctrlertypes "github.com/rigochain/rigo-go/ctrlers/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type TrxResult struct {
	*coretypes.ResultTx
	TxDetail *ctrlertypes.Trx `json:"tx_detail"`
}
