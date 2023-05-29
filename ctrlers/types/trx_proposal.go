package types

import (
	"bytes"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/rigochain/rigo-go/types/xerrors"
	"google.golang.org/protobuf/proto"
	"io"
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

func (tx *TrxPayloadProposal) Equal(_tx ITrxPayload) bool {
	if _tx == nil {
		return false
	}
	_tx0, ok := (_tx).(*TrxPayloadProposal)
	if !ok {
		return false
	}

	if tx.Message != _tx0.Message {
		return false
	}
	if tx.StartVotingHeight != _tx0.StartVotingHeight {
		return false
	}
	if tx.VotingPeriodBlocks != _tx0.VotingPeriodBlocks {
		return false
	}
	if tx.OptType != _tx0.OptType {
		return false
	}
	if len(tx.Options) != len(_tx0.Options) {
		return false
	}
	for i, d0 := range tx.Options {
		if bytes.Compare(d0, _tx0.Options[i]) != 0 {
			return false
		}
	}
	return true
}

func (tx *TrxPayloadProposal) Decode(bz []byte) xerrors.XError {
	pm := &TrxPayloadProposalProto{}
	if err := proto.Unmarshal(bz, pm); err != nil {
		return xerrors.From(err)
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
	return bz, xerrors.From(err)
}

func (tx *TrxPayloadProposal) RLPEncode() ([]byte, error) {
	return rlp.EncodeToBytes(tx)
}
func (tx *TrxPayloadProposal) RLPDecode(bz []byte) error {
	return rlp.DecodeBytes(bz, tx)
}

func (tx *TrxPayloadProposal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, tx)
}

func (tx *TrxPayloadProposal) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(tx)
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
