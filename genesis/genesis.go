package genesis

import (
	"encoding/binary"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/libs/crypto"
	"github.com/kysee/arcanus/types/account"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type GenesisAssetHolder struct {
	Address account.Address `json:"address"`
	Balance string          `json:"balance"`
}

func (gh *GenesisAssetHolder) Hash() []byte {
	hasher := crypto.DefaultHasher()
	hasher.Write(gh.Address[:])
	hasher.Write([]byte(gh.Balance))
	return hasher.Sum(nil)
}

type GenesisGovRule struct {
	Version            int64  `json:"version"`
	MaxValidatorCnt    int64  `json:"maxValidatorCnt"`
	AmountPerPower     string `json:"amountPerPower"`
	RewardPerPower     string `json:"rewardPerPower"`
	LazyRewardBlocks   int64  `json:"lazyRewardBlocks"`
	LazyApplyingBlocks int64  `json:"lazyApplyingBlocks"`
}

func DefaultGenesisGovRule() *GenesisGovRule {
	gr := gov.DefaultGovRule()
	return &GenesisGovRule{
		Version:            gr.Version,
		MaxValidatorCnt:    gr.MaxValidatorCnt,
		AmountPerPower:     gr.AmountPerPower.String(),
		RewardPerPower:     gr.RewardPerPower.String(),
		LazyApplyingBlocks: gr.LazyApplyingBlocks,
	}
}

func (gr *GenesisGovRule) Hash() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf, uint32(gr.Version))

	hasher := crypto.DefaultHasher()
	hasher.Write(buf)
	hasher.Write([]byte(gr.AmountPerPower))
	hasher.Write([]byte(gr.RewardPerPower))
	return hasher.Sum(nil)
}

type GenesisAppState struct {
	AssetHolders []*GenesisAssetHolder `json:"assetHolders"`
	GovRule      *GenesisGovRule       `json:"govRule"`
}

func (ga *GenesisAppState) Hash() ([]byte, error) {
	hasher := crypto.DefaultHasher()

	hasher.Write(ga.GovRule.Hash())
	for _, h := range ga.AssetHolders {
		hasher.Write(h.Hash())
	}

	return hasher.Sum(nil), nil
}

func NewGenesisDoc(chainID string, validators []tmtypes.GenesisValidator, assetHolders []*GenesisAssetHolder, govRule *GenesisGovRule) (*tmtypes.GenesisDoc, error) {
	appState := GenesisAppState{
		AssetHolders: assetHolders,
		GovRule:      govRule,
	}
	appStateJsonBlob, err := tmjson.Marshal(appState)
	if err != nil {
		return nil, err
	}
	appHash, err := appState.Hash()
	if err != nil {
		return nil, err
	}

	return &tmtypes.GenesisDoc{
		ChainID:     chainID,
		GenesisTime: tmtime.Now(),
		ConsensusParams: &tmproto.ConsensusParams{
			Block:    tmtypes.DefaultBlockParams(),
			Evidence: tmtypes.DefaultEvidenceParams(),
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeSecp256k1},
			},
			Version: tmproto.VersionParams{
				AppVersion: 1,
			},
		},
		Validators: validators,
		AppState:   appStateJsonBlob,
		AppHash:    appHash[:],
	}, nil
}
