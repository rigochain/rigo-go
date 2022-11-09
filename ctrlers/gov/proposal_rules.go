package gov

import (
	"github.com/kysee/arcanus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

type Vote struct {
	addr  types.Address
	power int64
}
type GovRulesProposal struct {
	TxHash            types.HexBytes   `json:"txHash"`
	StartVotingHeight int64            `json:"startVotingHeight"`
	LastVotingHeight  int64            `json:"lastVotingHeight"`
	ApplyingHeight    int64            `json:"applyingHeight"`
	MajorityPower     int64            `json:"majorityPower"`
	Votes             map[string]int64 `json:"votes"`
	Rules             []*GovRules      `json:"rules"`
}

func (p *GovRulesProposal) ID() []byte {
	return p.TxHash
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

func (p *GovRulesProposal) GetVotesFor(optidx int32) int64 {
	votes := int64(0)
	for _, v := range p.Votes {
		votes += v
	}
	return votes
}

func (p *GovRulesProposal) GetVotesOf(addr types.Address, optidx int32) int64 {
	v, ok := p.Votes[addr.String()]
	if !ok {
		v = 0
	}
	return v
}

func (p *GovRulesProposal) DoVote(addr types.Address, power int64) {
	if power > 0 {
		p.Votes[addr.String()] = power
	}
}

var _ types.IProposable = (*GovRulesProposal)(nil)
