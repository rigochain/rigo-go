package genesis

import (
	types2 "github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/rigochain/rigo-go/types"
	"github.com/rigochain/rigo-go/types/crypto"
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
	hasher := crypto.DefaultHasher()
	hasher.Write(gh.Address[:])
	hasher.Write([]byte(gh.Balance))
	return hasher.Sum(nil)
}

type GenesisAppState struct {
	AssetHolders []*GenesisAssetHolder `json:"assetHolders"`
	GovRule      *types2.GovRule       `json:"govRule"`
}

func (ga *GenesisAppState) Hash() ([]byte, error) {
	hasher := crypto.DefaultHasher()
	if bz, err := ga.GovRule.Encode(); err != nil {
		return nil, err
	} else if _, err := hasher.Write(bz); err != nil {
		return nil, err
	} else {
		for _, h := range ga.AssetHolders {
			if _, err := hasher.Write(h.Hash()); err != nil {
				return nil, err
			}
		}
	}
	return hasher.Sum(nil), nil
}

func NewGenesisDoc(chainID string, validators []tmtypes.GenesisValidator, assetHolders []*GenesisAssetHolder, govRule *types2.GovRule) (*tmtypes.GenesisDoc, error) {
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
