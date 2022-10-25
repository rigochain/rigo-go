package genesis

import (
	"fmt"
	"github.com/kysee/arcanus/libs"
	"github.com/stretchr/testify/require"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/rand"
	"testing"
)

func TestNewGenesis(t *testing.T) {
	holders := make([]*GenesisAssetHolder, 100)
	for i, _ := range holders {
		holders[i] = &GenesisAssetHolder{
			Address: libs.RandAddress(),
			Balance: fmt.Sprintf("%v", rand.Int63()),
		}
	}

	validators := make([]tmtypes.GenesisValidator, 10)
	for i, _ := range validators {
		validators[i] = tmtypes.GenesisValidator{
			Address: libs.RandAddress(),
			PubKey:  tmsecp256k1.PubKey(libs.RandBytes(33)),
			Power:   rand.Int63n(10000),
			Name:    fmt.Sprintf("test-validator #%v", i),
		}
	}

	govRules := DefaultGenesisGovRules()
	genDoc, err := NewGenesisDoc("TestChainID", validators, holders, govRules)
	require.NoError(t, err)

	bzJson, err := tmjson.MarshalIndent(genDoc, "", "   ")
	require.NoError(t, err)
	fmt.Println(string(bzJson))
}

func TestDevnetGenesisUnmarshal(t *testing.T) {
	genDoc := &tmtypes.GenesisDoc{}
	err := tmjson.Unmarshal(jsonBlobDevnetGenesis, genDoc)
	require.NoError(t, err)

	appState := &GenesisAppState{}
	err = tmjson.Unmarshal(genDoc.AppState, appState)
	require.NoError(t, err)

	require.Equal(t, int64(0), appState.GovRules.Version)
	require.Equal(t, "1000000000000000000", appState.GovRules.AmountPerPower)
	require.Equal(t, "1000000000", appState.GovRules.RewardPerPower)
}
