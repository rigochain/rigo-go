package gov

import (
	"github.com/kysee/arcanus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

type Vote struct {
	addr  types.Address
	power int64
}
type GovRuleProposal struct {
	TxHash            types.HexBytes   `json:"txHash"`
	StartVotingHeight int64            `json:"startVotingHeight"`
	LastVotingHeight  int64            `json:"lastVotingHeight"`
	ApplyingHeight    int64            `json:"applyingHeight"`
	MajorityPower     int64            `json:"majorityPower"`
	Votes             map[string]int64 `json:"votes"`
	Rules             []*GovRule       `json:"rules"`
}

func (p *GovRuleProposal) ID() []byte {
	return p.TxHash
}

func (p *GovRuleProposal) Type() int32 {
	return types.PROPOSAL_GOVRULE
}

func (p *GovRuleProposal) Encode() ([]byte, error) {
	return tmjson.Marshal(p)
}

func (p *GovRuleProposal) Decode(d []byte) error {
	return tmjson.Unmarshal(d, p)
}

func (p *GovRuleProposal) GetStartVotingHeight() int64 {
	return p.StartVotingHeight
}

func (p *GovRuleProposal) GetLastVotingHeight() int64 {
	return p.LastVotingHeight
}

func (p *GovRuleProposal) GetApplyingHeight() int64 {
	return p.ApplyingHeight
}

func (p *GovRuleProposal) GetMajorityPower() int64 {
	return p.MajorityPower
}

func (p *GovRuleProposal) GetOption(i int) interface{} {
	return p.Rules[i]
}

func (p *GovRuleProposal) GetVotesFor(optidx int32) int64 {
	votes := int64(0)
	for _, v := range p.Votes {
		votes += v
	}
	return votes
}

func (p *GovRuleProposal) GetVotesOf(addr types.Address, optidx int32) int64 {
	v, ok := p.Votes[addr.String()]
	if !ok {
		v = 0
	}
	return v
}

func (p *GovRuleProposal) DoVote(addr types.Address, power int64) {
	if power > 0 {
		p.Votes[addr.String()] = power
	}
}

var _ types.IProposable = (*GovRuleProposal)(nil)
