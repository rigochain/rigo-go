package web3

import (
	"github.com/rigochain/rigo-go/libs/web3/types"
	"sync"
)

type RigoWeb3 struct {
	provider types.Provider
	callId   int64
	mtx      sync.RWMutex
}

func NewRigoWeb3(provider types.Provider) *RigoWeb3 {
	return &RigoWeb3{
		provider: provider,
	}
}

func (rweb3 *RigoWeb3) NewRequest(method string, args ...interface{}) (*types.JSONRpcReq, error) {
	rweb3.mtx.Lock()
	defer rweb3.mtx.Unlock()

	rweb3.callId++

	return types.NewRequest(rweb3.callId, method, args...)
}
