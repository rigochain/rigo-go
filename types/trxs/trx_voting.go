package trxs

import (
	"github.com/golang/protobuf/proto"
	"github.com/kysee/arcanus/types"
)

type TrxPayloadVoting struct {
	TxHash types.HexBytes
	Choice int32
}

var _ ITrxPayload = (*TrxPayloadVoting)(nil)

func (tx *TrxPayloadVoting) Type() int32 {
	return TRX_VOTING
}

func (tx *TrxPayloadVoting) Decode(bz []byte) error {
	pm := &TrxPayloadVotingProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return err
	}

	tx.TxHash = pm.TxHash
	tx.Choice = pm.Choice
	return nil
}

func (tx *TrxPayloadVoting) Encode() ([]byte, error) {
	pm := &TrxPayloadVotingProto{
		TxHash: tx.TxHash,
		Choice: tx.Choice,
	}

	return proto.Marshal(pm)
}
