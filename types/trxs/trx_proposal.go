package trxs

import (
	"github.com/golang/protobuf/proto"
)

type TrxPayloadProposal struct {
	Message            string
	VotingBlocks       int64
	LazyApplyingBlocks int64
	ProposalType       int32
	Options            [][]byte
}

var _ ITrxPayload = (*TrxPayloadProposal)(nil)

func (tx *TrxPayloadProposal) Type() int32 {
	return TRX_PROPOSAL
}

func (tx *TrxPayloadProposal) Decode(bz []byte) error {
	pm := &TrxPayloadProposalProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return err
	}

	tx.Message = pm.Message
	tx.VotingBlocks = pm.VotingBlocks
	tx.LazyApplyingBlocks = pm.LazyApplyingBlocks
	tx.ProposalType = pm.ProposalType
	tx.Options = pm.Options
	return nil
}

func (tx *TrxPayloadProposal) Encode() ([]byte, error) {
	pm := &TrxPayloadProposalProto{
		Message:            tx.Message,
		VotingBlocks:       tx.VotingBlocks,
		LazyApplyingBlocks: tx.LazyApplyingBlocks,
		ProposalType:       tx.ProposalType,
		Options:            tx.Options,
	}

	return proto.Marshal(pm)
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
