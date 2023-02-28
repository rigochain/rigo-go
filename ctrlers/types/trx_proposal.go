package types

import (
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
)

type TrxPayloadProposal struct {
	Message            string
	StartVotingHeight  int64
	VotingPeriodBlocks int64
	OptType            int32
	Options            [][]byte
}

var _ ITrxPayload = (*TrxPayloadProposal)(nil)

func (tx *TrxPayloadProposal) Type() int32 {
	return TRX_PROPOSAL
}

func (tx *TrxPayloadProposal) Decode(bz []byte) xerrors.XError {
	pm := &TrxPayloadProposalProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.NewFrom(err)
	}

	tx.Message = pm.Message
	tx.StartVotingHeight = pm.StartVotingHeight
	tx.VotingPeriodBlocks = pm.VotingBlocks
	tx.OptType = pm.OptType
	tx.Options = pm.Options
	return nil
}

func (tx *TrxPayloadProposal) Encode() ([]byte, xerrors.XError) {
	pm := &TrxPayloadProposalProto{
		Message:           tx.Message,
		StartVotingHeight: tx.StartVotingHeight,
		VotingBlocks:      tx.VotingPeriodBlocks,
		OptType:           tx.OptType,
		Options:           tx.Options,
	}

	bz, err := proto.Marshal(pm)
	return bz, xerrors.NewFrom(err)
}

func (tx *TrxPayloadProposal) GetOptionCount() int {
	return len(tx.Options)
}

func (tx *TrxPayloadProposal) GetOption(i int) []byte {
	if i >= len(tx.Options) {
		return nil
	}
	return tx.Options[i]
}
