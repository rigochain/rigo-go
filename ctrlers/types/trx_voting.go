package types

import (
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
)

type TrxPayloadVoting struct {
	TxHash bytes.HexBytes
	Choice int32
}

var _ ITrxPayload = (*TrxPayloadVoting)(nil)

func (tx *TrxPayloadVoting) Type() int32 {
	return TRX_VOTING
}

func (tx *TrxPayloadVoting) Decode(bz []byte) xerrors.XError {
	pm := &TrxPayloadVotingProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.NewFrom(err)
	}

	tx.TxHash = pm.TxHash
	tx.Choice = pm.Choice
	return nil
}

func (tx *TrxPayloadVoting) Encode() ([]byte, xerrors.XError) {
	pm := &TrxPayloadVotingProto{
		TxHash: tx.TxHash,
		Choice: tx.Choice,
	}

	bz, err := proto.Marshal(pm)
	return bz, xerrors.NewFrom(err)
}
