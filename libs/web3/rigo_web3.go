package web3

import (
	"github.com/rigochain/rigo-go/libs/web3/types"
	"sync"
)

type RigoWeb3 struct {
	chainId  string
	provider types.Provider
	callId   int64
	mtx      sync.RWMutex
}

func NewRigoWeb3(provider types.Provider) *RigoWeb3 {
	types.NewRequest(0, "genesis")

	rweb3 := &RigoWeb3{
		provider: provider,
	}
	gen, err := rweb3.Genesis()
	if err != nil {
		return nil
	}
	rweb3.chainId = gen.Genesis.ChainID
	return rweb3
}

func (rweb3 *RigoWeb3) ChainID() string {
	rweb3.mtx.RLock()
	defer rweb3.mtx.RUnlock()

	return rweb3.chainId
}

func (rweb3 *RigoWeb3) SetChainID(cid string) {
	rweb3.mtx.RLock()
	defer rweb3.mtx.RUnlock()

	rweb3.chainId = cid
}

func (rweb3 *RigoWeb3) NewRequest(method string, args ...interface{}) (*types.JSONRpcReq, error) {
	rweb3.mtx.Lock()
	defer rweb3.mtx.Unlock()

	rweb3.callId++

	return types.NewRequest(rweb3.callId, method, args...)
}
