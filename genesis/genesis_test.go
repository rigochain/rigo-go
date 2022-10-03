package genesis

import (
	"fmt"
	"github.com/kysee/arcanus/libs"
	"github.com/stretchr/testify/require"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/rand"
	"testing"
)

func TestNewGenesis(t *testing.T) {
	holders := make([]GenesisAssetHolder, 100)
	for i, _ := range holders {
		holders[i] = GenesisAssetHolder{
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

	genDoc, err := NewGenesisDoc("TestChainID", validators, holders)
	require.NoError(t, err)

	bzJson, err := json.MarshalIndent(genDoc, "", "   ")
	require.NoError(t, err)
	fmt.Println(string(bzJson))
}
