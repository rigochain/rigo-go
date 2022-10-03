package genesis

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

func MainnetGenesisDoc() (*tmtypes.GenesisDoc, error) {
	return DevnetGenesisDoc()
}
