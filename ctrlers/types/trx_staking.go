package types

import (
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
)

type TrxPayloadStaking struct{}

func (tx *TrxPayloadStaking) Type() int32 {
	return TRX_STAKING
}

func (tx *TrxPayloadStaking) Decode(bz []byte) xerrors.XError {
	return nil
}

func (tx *TrxPayloadStaking) Encode() ([]byte, xerrors.XError) {
	return nil, nil
}

type TrxPayloadUnstaking struct {
	TxHash bytes.HexBytes `json:"txhash"`
}

func (tx *TrxPayloadUnstaking) Type() int32 {
	return TRX_UNSTAKING
}

func (tx *TrxPayloadUnstaking) Decode(bz []byte) xerrors.XError {
	pm := &TrxPayloadUnstakingProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.From(err)
	}
	tx.TxHash = pm.TxHash
	return nil
}

func (tx *TrxPayloadUnstaking) Encode() ([]byte, xerrors.XError) {
	pm := &TrxPayloadUnstakingProto{
		TxHash: tx.TxHash,
	}

	bz, err := proto.Marshal(pm)
	return bz, xerrors.From(err)
}
