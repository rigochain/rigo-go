package genesis

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/kysee/arcanus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type GenesisAssetHolder struct {
	Address types.Address `json:"address"`
	Balance string        `json:"balance"`
}

func (gh *GenesisAssetHolder) Hash() []byte {
	hasher := sha256.New()
	hasher.Write(gh.Address[:])
	hasher.Write([]byte(gh.Balance))
	return hasher.Sum(nil)
}

type GenesisGovRules struct {
	Version        uint64 `json:"version"`
	AmountPerPower string `json:"amountPerPower"`
	RewardPerPower string `json:"rewardPerPower"`
}

func DefaultGenesisGovRules() *GenesisGovRules {
	return &GenesisGovRules{
		Version:        0,
		AmountPerPower: "1000000000000000000", // 1_000000000000000000
		RewardPerPower: "1000000000",
	}
}

func (gr *GenesisGovRules) Hash() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, gr.Version)

	hasher := sha256.New()
	hasher.Write(buf)
	hasher.Write([]byte(gr.AmountPerPower))
	hasher.Write([]byte(gr.RewardPerPower))
	return hasher.Sum(nil)
}

type GenesisAppState struct {
	AssetHolders []*GenesisAssetHolder `json:"assetHolders"`
	GovRules     *GenesisGovRules      `json:"govRules"`
}

func (ga *GenesisAppState) Hash() ([]byte, error) {
	hasher := sha256.New()

	hasher.Write(ga.GovRules.Hash())
	for _, h := range ga.AssetHolders {
		hasher.Write(h.Hash())
	}

	return hasher.Sum(nil), nil
}

func NewGenesisDoc(chainID string, validators []tmtypes.GenesisValidator, assetHolders []*GenesisAssetHolder, govRules *GenesisGovRules) (*tmtypes.GenesisDoc, error) {
	appState := GenesisAppState{
		AssetHolders: assetHolders,
		GovRules:     govRules,
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
