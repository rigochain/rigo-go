package rpc

import (
	"encoding/json"
	"github.com/kysee/arcanus/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

type QueryResponse struct {
	//it's from abcitypes.ResponseQuery
	Code      uint32           `json:"code,omitempty"`
	Log       string           `json:"log,omitempty"`
	Info      string           ` json:"info,omitempty"`
	Index     int64            `json:"index,omitempty"`
	Key       types.HexBytes   `json:"key,omitempty"`
	Value     json.RawMessage  `json:"value,omitempty"`
	ProofOps  *crypto.ProofOps `json:"proof_ops,omitempty"`
	Height    int64            `json:"height,omitempty"`
	Codespace string           `json:"codespace,omitempty"`
}

func ToQueryResponse(r *abcitypes.ResponseQuery) *QueryResponse {
	return &QueryResponse{
		Code:      r.Code,
		Log:       r.Log,
		Info:      r.Info,
		Index:     r.Index,
		Key:       r.Key,
		Value:     r.Value,
		ProofOps:  r.ProofOps,
		Height:    r.Height,
		Codespace: r.Codespace,
	}
}
