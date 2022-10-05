package gov

import (
	"github.com/kysee/arcanus/types"
	"github.com/kysee/arcanus/types/xerrors"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
)

type GovCtrler struct {
	govDB  db.DB
	rules  types.IGovRules
	logger log.Logger
}

func NewGovCtrler(dbDir string, logger log.Logger) (*GovCtrler, error) {
	govDB, err := db.NewDB("gov", "goleveldb", dbDir)
	if err != nil {
		return nil, err
	}

	return &GovCtrler{govDB: govDB, logger: logger}, nil
}

func (ctrler *GovCtrler) SetRules(rules types.IGovRules) {
	ctrler.rules = rules
}

func (ctrler *GovCtrler) GetRules() types.IGovRules {
	return ctrler.rules
}

func (ctrler *GovCtrler) ImportRules(cb func() []byte) error {
	bz := cb()
	if bz == nil {
		return xerrors.New("rule blob is nil")
	} else if rules, err := DecodeGovRules(bz); err != nil {
		return xerrors.Wrap(err)
	} else {
		ctrler.rules = rules
	}
	return nil
}
