package proposals

import (
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

type GovRulesProposal struct {
	id                int64           `json:"id"`
	StartVotingHeight int64           `json:"startVotingHeight"`
	LastVotingHeight  int64           `json:"lastVotingHeight"`
	ApplyingHeight    int64           `json:"applyingHeight"`
	MajorityPower     int64           `json:"majorityPower"`
	Rules             []*gov.GovRules `json:"rules"`
	Votes             []int64         `json:"votes"`
}

func (p *GovRulesProposal) ID() int64 {
	return p.id
}

func (p *GovRulesProposal) Type() int32 {
	return types.PROPOSAL_GOVRULES
}

func (p *GovRulesProposal) Encode() ([]byte, error) {
	return tmjson.Marshal(p)
}

func (p *GovRulesProposal) Decode(d []byte) error {
	return tmjson.Unmarshal(d, p)
}

func (p *GovRulesProposal) GetStartVotingHeight() int64 {
	return p.StartVotingHeight
}

func (p *GovRulesProposal) GetLastVotingHeight() int64 {
	return p.LastVotingHeight
}

func (p *GovRulesProposal) GetApplyingHeight() int64 {
	return p.ApplyingHeight
}

func (p *GovRulesProposal) GetMajorityPower() int64 {
	return p.MajorityPower
}

func (p *GovRulesProposal) GetOption(i int) interface{} {
	return p.Rules[i]
}

func (p *GovRulesProposal) GetVote(i int) int64 {
	return p.Votes[i]
}

var _ types.IProposable = (*GovRulesProposal)(nil)
