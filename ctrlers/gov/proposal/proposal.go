package proposal

import (
	"encoding/json"
	"github.com/kysee/arcanus/ledger"
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/bytes"
	"github.com/kysee/arcanus/types/xerrors"
	"sort"
	"sync"
)

type GovProposal struct {
	GovProposalHeader `json:"header"`
	Options           []*voteOption `json:"options"`
	MajorOption       *voteOption   `json:"majorOption"`

	mtx sync.RWMutex
}

func NewGovProposal(txhash bytes.HexBytes, optType int32, startHeight, votingBlocks, lazyApplyingBlocks, totalVotingPower int64, voters map[string]*Voter, options ...[]byte) *GovProposal {
	return &GovProposal{
		GovProposalHeader: GovProposalHeader{
			TxHash:            txhash,
			StartVotingHeight: startHeight,
			EndVotingHeight:   startHeight + votingBlocks,
			ApplyingHeight:    startHeight + votingBlocks + lazyApplyingBlocks,
			TotalVotingPower:  totalVotingPower,
			MajorityPower:     totalVotingPower * 2 / 3,
			Voters:            voters,
			OptType:           optType,
		},
		Options:     NewVoteOptions(options...),
		MajorOption: nil,
	}
}

func (prop *GovProposal) Key() ledger.LedgerKey {
	prop.mtx.RLock()
	defer prop.mtx.RUnlock()

	return prop.TxHash.Array32()
}

func (prop *GovProposal) Encode() ([]byte, xerrors.XError) {
	prop.mtx.RLock()
	defer prop.mtx.RUnlock()

	if bz, err := json.Marshal(prop); err != nil {
		return bz, xerrors.NewFrom(err)
	} else {
		return bz, nil
	}
}

func (prop *GovProposal) Decode(bz []byte) xerrors.XError {
	prop.mtx.Lock()
	defer prop.mtx.Unlock()

	if err := json.Unmarshal(bz, prop); err != nil {
		return xerrors.NewFrom(err)
	}
	return nil
}

var _ ledger.ILedgerItem = (*GovProposal)(nil)

func (prop *GovProposal) DoVote(addr types.Address, choice int32) xerrors.XError {
	prop.mtx.Lock()
	defer prop.mtx.Unlock()

	// cancel previous vote
	voter := prop.Voters[addr.String()]
	if voter == nil {
		return xerrors.New("not found voter")
	}

	prop.cancelVote(voter)
	prop.doVote(voter, choice)

	return nil
}

func (prop *GovProposal) cancelVote(voter *Voter) {
	if voter.Choice >= 0 {
		opt := prop.Options[voter.Choice]
		opt.CancelVote(voter.Power)
		voter.Choice = -1
	}
}

func (prop *GovProposal) doVote(voter *Voter, choice int32) {
	if choice >= 0 {
		opt := prop.Options[choice]
		if opt == nil {
			return //xerrors.New("not found option")
		}

		opt.DoVote(voter.Power)
		voter.Choice = choice
	}
}

func (prop *GovProposal) UpdateMajorOption() *voteOption {
	prop.mtx.Lock()
	defer prop.mtx.Unlock()

	return prop.updateMajorOption()
}

func (prop *GovProposal) updateMajorOption() *voteOption {
	sort.Sort(powerOrderVoteOptions(prop.Options))
	if prop.Options[0].Votes() >= prop.MajorityPower {
		prop.MajorOption = prop.Options[0]
	}
	return prop.MajorOption
}

func (prop *GovProposal) isMajor(opt *voteOption) bool {
	return opt.Votes() >= prop.MajorityPower
}

func (prop *GovProposal) IsVoter(addr types.Address) bool {
	_, ok := prop.Voters[addr.String()]
	return ok
}

type powerOrderVoteOptions []*voteOption

func (opts powerOrderVoteOptions) Len() int {
	return len(opts)
}

func (opts powerOrderVoteOptions) Less(i, j int) bool {
	return opts[i].votes > opts[j].votes
}

func (opts powerOrderVoteOptions) Swap(i, j int) {
	opts[i], opts[j] = opts[j], opts[i]
}

var _ sort.Interface = (powerOrderVoteOptions)(nil)
