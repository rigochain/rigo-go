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
