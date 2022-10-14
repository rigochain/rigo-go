package trxs

import (
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
)

type TrxPayloadStaking struct{}

func (tx *TrxPayloadStaking) Type() int32 {
	return TRX_STAKING
}

func (tx *TrxPayloadStaking) Decode(bz []byte) error {
	return nil
}

func (tx *TrxPayloadStaking) Encode() ([]byte, error) {
	return nil, nil
}

type TrxPayloadUnstaking struct {
	TxHash types.HexBytes `json:"txhash"`
}

func (tx *TrxPayloadUnstaking) Type() int32 {
	return TRX_UNSTAKING
}

func (tx *TrxPayloadUnstaking) Decode(bz []byte) error {
	pm := &TrxPayloadUnstakingProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return err
	}
	tx.TxHash = pm.TxHash
	return nil
}

func (tx *TrxPayloadUnstaking) Encode() ([]byte, error) {
	pm := &TrxPayloadUnstakingProto{
		TxHash: tx.TxHash,
	}
	return proto.Marshal(pm)
}
