package types

import (
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
)

type TrxPayloadContract struct {
	Data []byte `json:"data"`
}

func (tx *TrxPayloadContract) Type() int32 {
	return TRX_CONTRACT
}

func (tx *TrxPayloadContract) Encode() ([]byte, xerrors.XError) {
	pm := &TrxPayloadContractProto{
		XData: tx.Data,
	}

	bz, err := proto.Marshal(pm)
	return bz, xerrors.From(err)
}

func (tx *TrxPayloadContract) Decode(bz []byte) xerrors.XError {
	pm := &TrxPayloadContractProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.From(err)
	}

	tx.Data = pm.XData
	return nil
}

var _ ITrxPayload = (*TrxPayloadContract)(nil)
