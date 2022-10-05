package genesis

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/kysee/arcanus/ctrlers/gov"
	"github.com/kysee/arcanus/types"
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

type GenesisAppState struct {
	AssetHolders []GenesisAssetHolder `json:"assetHolders"`
	GovRules     *gov.GovRules        `json:"govRules"`
}

func (ga *GenesisAppState) Hash() ([]byte, error) {
	hasher := sha256.New()

	bzgr, err := ga.GovRules.Encode()
	if err != nil {
		return nil, err
	} else {
		hasher.Write(bzgr)
	}

	for _, h := range ga.AssetHolders {
		hasher.Write(h.Hash())
	}

	return hasher.Sum(nil), nil
}

func NewGenesisDoc(chainID string, validators []tmtypes.GenesisValidator, assetHolders []GenesisAssetHolder) (*tmtypes.GenesisDoc, error) {
	appState := GenesisAppState{
		AssetHolders: assetHolders,
	}
	appStateJsonBlob, err := json.Marshal(appState)
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
