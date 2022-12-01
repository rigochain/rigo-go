package state

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	tmproxy "github.com/tendermint/tendermint/proxy"
)

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type arcanusLocalClientCreator struct {
	mtx *tmsync.Mutex
	app abcitypes.Application
}

// NewLocalClientCreator returns a ClientCreator for the given app,
// which will be running locally.
func NewArcanusLocalClientCreator(app abcitypes.Application) tmproxy.ClientCreator {
	return &arcanusLocalClientCreator{
		mtx: new(tmsync.Mutex),
		app: app,
	}
}

func (l *arcanusLocalClientCreator) NewABCIClient() (abcicli.Client, error) {
	client := abcicli.NewLocalClient(l.mtx, l.app)
	l.app.(*ChainCtrler).SetAppConnConsensus(client)
	return client, nil
}
