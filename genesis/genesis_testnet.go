package genesis

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestnetGenesisDoc() (*tmtypes.GenesisDoc, error) {
	return DevnetGenesisDoc()
}
